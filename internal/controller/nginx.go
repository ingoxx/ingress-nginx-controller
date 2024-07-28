package controller

import (
	"bytes"
	"context"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/controller/store"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
	"text/template"
)

const (
	nginxTmpl   = "/rootfs/etc/nginx/template/nginx.tmpl"
	serverTmpl  = "/rootfs/etc/nginx/template/server.tmpl"
	sslPath     = "/etc/nginx/ssl"
	testConfDir = "/etc/nginx/conf.d"
)

type configure struct {
	Server      *ingressv1.Server
	Annotations *annotations.Ingress
	ServerTpl   bytes.Buffer
}

type NginxConfigure struct {
	client  client.Client
	ctx     context.Context
	rr      resolver.Resolver
	mux     *sync.RWMutex
	ingress *ingressv1.Ingress
}

func NewNginxConfigure(store store.Storer) *NginxConfigure {
	st := store.GetReconcilerInfo()
	n := &NginxConfigure{
		client:  st.Client,
		ctx:     st.Context,
		rr:      st.IngressInfos,
		ingress: st.Ingress,
		mux:     new(sync.RWMutex),
	}
	return n
}

func (n *NginxConfigure) generateServer(config *configure) error {
	serverStr, err := os.ReadFile(serverTmpl)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file: %s not found", serverTmpl))
		return err
	}

	serverTemp, err := template.New("serverMain").Parse(string(serverStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template: %s", serverTmpl))
		return err
	}

	if err := serverTemp.Execute(&config.ServerTpl, config); err != nil {
		return err
	}

	return nil
}

func (n *NginxConfigure) generateNginxConfigure(cfg *ingressv1.Configuration, annotations *annotations.Ingress) error {
	mainTmplStr, err := os.ReadFile(nginxTmpl)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file: %s not found", nginxTmpl))
		return err
	}

	mainTmpl, err := template.New("main").Parse(string(mainTmplStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template: %s", nginxTmpl))
		return err
	}

	var config = new(configure)
	config.Annotations = annotations
	if cfg != nil {
		for _, v := range cfg.Servers {
			config.Server = v
			if err := n.generateServer(config); err != nil {
				klog.ErrorS(err, "fail to generate server template")
				return err
			}
		}
	}

	//生成的server块插入到指定位置
	_, err = mainTmpl.New("servers").Parse(config.ServerTpl.String())
	if err != nil {
		return err
	}

	// 执行渲染
	var tpl bytes.Buffer
	err = mainTmpl.Execute(&tpl, nil)
	if err != nil {
		return err
	}

	if err := n.generateTestConf(tpl.Bytes()); err != nil {
		return err
	}

	return nil
}

func (n *NginxConfigure) generateTestConf(b []byte) error {
	if err := os.WriteFile(filepath.Join(testConfDir, n.ingress.Name+"-test.conf"), b, 0644); err != nil {
		klog.ErrorS(err, fmt.Sprintf("an error occurred while writing the generated template content to %s", testConfDir))
		return err
	}

	return nil
}

func (n *NginxConfigure) GenerateNginxConfigure(ingress annotations.IngressAnnotations) error {
	n.mux.Lock()
	defer n.mux.Unlock()
	cfg := n.getBackendConfigure(ingress)
	if ingress.Ingress.Spec.DefaultBackend != nil {
		defaultServer, err := n.getDefaultBackendConfigure(ingress)
		if err == nil {
			cfg.Servers = append(cfg.Servers, defaultServer)
		}
	}

	if cfg == nil {
		return fmt.Errorf("fail to get backend config in ingress: %s, namespace: %s", ingress.Ingress.Name, ingress.Ingress.Namespace)
	}

	if err := n.generateNginxConfigure(cfg, ingress.ParsedAnnotations); err != nil {
		return err
	}

	klog.Info("update nginx configure successfully")

	//if err := nginx.Reload(nginxConf, testConf); err != nil {
	//	return err
	//}

	return nil
}

func (n *NginxConfigure) getDefaultBackendConfigure(ingress annotations.IngressAnnotations) (*ingressv1.Server, error) {
	var servers *ingressv1.Server
	var backends []*ingressv1.Backend

	svc, err := n.rr.GetService(ingress.Ingress.Spec.DefaultBackend.Service.Name)
	if err != nil {
		return servers, err
	}

	backendPort := n.rr.GetSvcPort(*ingress.Ingress.Spec.DefaultBackend)
	if backendPort == 0 {
		klog.ErrorS(fmt.Errorf("%s svc port not exists", svc.Name), fmt.Sprintf("namespace: %s", ingress.Ingress.Namespace))
		return servers, fmt.Errorf("%s svc port not exists", svc.Name)
	}

	b := &ingressv1.Backend{
		Name:           svc.Name,
		NameSpace:      svc.Namespace,
		Port:           backendPort,
		Path:           "/",
		ServiceBackend: ingress.Ingress.Spec.DefaultBackend.Service,
	}
	backends = append(backends, b)

	s := &ingressv1.Server{
		Name:      ingress.Ingress.Name,
		NameSpace: ingress.Ingress.Namespace,
		HostName:  "_",
		Paths:     backends,
	}

	return s, nil
}

func (n *NginxConfigure) getBackendConfigure(ingress annotations.IngressAnnotations) *ingressv1.Configuration {
	var rule = ingress.Ingress.Spec.Rules
	var servers = make([]*ingressv1.Server, len(rule))

	tls, err := n.generateTlsFile(ingress)
	if err != nil {
		tls.TlsNoPass = false
	}

	for k, v := range rule {
		var backend = make([]*ingressv1.Backend, len(v.HTTP.Paths))
		for bk, p := range v.HTTP.Paths {
			if ingress.ParsedAnnotations.Rewrite.RewriteEnableRegex {
				if *p.PathType != "ImplementationSpecific" {
					klog.ErrorS(
						fmt.Errorf("when annotations rewrite-enable-regex is true, the value of pathType must be: ImplementationSpecific"),
						fmt.Sprintf("ingress: %s, namespace: %s", ingress.Ingress.Name, ingress.Ingress.Namespace))
					return nil
				}
			}

			svc, err := n.rr.GetService(p.Backend.Service.Name)
			if err != nil {
				return nil
			}

			backendPort := n.rr.GetSvcPort(p.Backend)
			if backendPort == 0 {
				klog.ErrorS(fmt.Errorf("%s svc port not exists", p.Backend.Service.Name), fmt.Sprintf("not found, svc name: %s, namespace: %s", p.Backend.Service.Name, ingress.Ingress.Namespace))
				return nil
			}

			b := &ingressv1.Backend{
				Name:           svc.Name,
				NameSpace:      svc.Namespace,
				Path:           p.Path,
				Port:           backendPort,
				ServiceBackend: p.Backend.Service,
				Annotations:    ingress.ParsedAnnotations,
			}
			backend = append(backend[:bk], b)
		}

		s := &ingressv1.Server{
			Name:      ingress.Ingress.Name,
			NameSpace: ingress.Ingress.Namespace,
			HostName:  v.Host,
			Paths:     backend,
			Tls:       tls,
		}

		servers = append(servers[:k], s)
	}

	return &ingressv1.Configuration{Servers: servers}
}

func (n *NginxConfigure) generateTlsFile(ingress annotations.IngressAnnotations) (ingressv1.SSLCert, error) {
	var ssl = ingressv1.SSLCert{
		TlsNoPass: true,
	}
	key := types.NamespacedName{Name: ingress.Ingress.Name + "-secret", Namespace: ingress.Ingress.Namespace}
	data, err := n.rr.GetTlsData(key)
	if err != nil {
		return ssl, err
	}

	for k, v := range data {
		file := filepath.Join(sslPath, k)
		if err := os.WriteFile(file, v, 0644); err != nil {
			klog.ErrorS(err, fmt.Sprintf("an error occurred while writing the generated template content to %s.", file))
			return ssl, err
		}
		if k == "tls.crt" {
			ssl.TlsCrt = file
		} else if k == "tls.key" {
			ssl.TlsKey = file
		}
	}

	return ssl, nil
}
