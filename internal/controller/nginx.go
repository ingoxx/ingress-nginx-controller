package controller

import (
	"bytes"
	"context"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/config"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/controller/store"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/nginx"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync"
	"text/template"
)

type configure struct {
	Server      *ingressv1.Server
	Annotations *annotations.Ingress
	ServerTpl   bytes.Buffer
	Cfg         *ingressv1.Configuration
	TmplName    string
	MainTmpl    string
	ConfName    string
}

type NginxController struct {
	client  client.Client
	ctx     context.Context
	rr      resolver.Resolver
	mux     *sync.RWMutex
	ingress *ingressv1.Ingress
}

func NewNginxController(store store.Storer) *NginxController {
	st := store.ReconcilerInfo()
	n := &NginxController{
		client:  st.Client,
		ctx:     st.Context,
		rr:      st.IngressInfos,
		ingress: st.Ingress,
		mux:     new(sync.RWMutex),
	}

	return n
}

func (n *NginxController) generateServerBytes(cfg *configure) error {
	serverStr, err := os.ReadFile(cfg.TmplName)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file: %s not found", cfg.TmplName))
		return err
	}

	serverTemp, err := template.New("serverMain").Parse(string(serverStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s", cfg.TmplName))
		return err
	}

	if err := serverTemp.Execute(&cfg.ServerTpl, cfg); err != nil {
		return err
	}

	return nil
}

func (n *NginxController) generateConf(name string, b []byte) error {
	testConf := name + "-test.conf"
	if err := os.WriteFile(testConf, b, 0644); err != nil {
		klog.ErrorS(err, fmt.Sprintf("an error occurred while writing the generated content to %s", testConf))
		return err
	}

	stat, err := os.Stat(testConf)
	if err != nil || stat.Size() == 0 {
		klog.ErrorS(err, "fail to generate file")
		return err
	}

	return nil
}

// Generate a.conf file named after host
func (n *NginxController) generateConfigureBytes(cfg *configure) error {
	mainTmplStr, err := os.ReadFile(cfg.MainTmpl)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("tmpelate file: %s not found", cfg.MainTmpl))
		return err
	}

	mainTmpl, err := template.New("main").Parse(string(mainTmplStr))
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("error parsing template_nginx: %s", cfg.MainTmpl))
		return err
	}

	if cfg != nil {
		for _, v := range cfg.Cfg.Servers {
			cfg.Server = v
			if err := n.generateServerBytes(cfg); err != nil {
				klog.ErrorS(err, "fail to generate server template_nginx")
				return err
			}

		}
	}

	_, err = mainTmpl.New("servers").Parse(cfg.ServerTpl.String())
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err = mainTmpl.Execute(&tpl, nil); err != nil {
		return err
	}

	if err := n.generateConf(cfg.ConfName, tpl.Bytes()); err != nil {
		return err
	}

	return nil
}

func (n *NginxController) GenerateConfigure(ingress annotations.IngressAnnotations) error {
	n.mux.Lock()
	defer n.mux.Unlock()

	if len(n.ingress.Spec.Rules) > 0 {
		if err := n.generateBackendTemplate(ingress); err != nil {
			return err
		}
	}

	if n.ingress.Spec.DefaultBackend != nil {
		if err := n.generateDefaultBackendTemplate(ingress); err != nil {
			return err
		}
	}

	return nil
}

func (n *NginxController) generateBackendTemplate(ingress annotations.IngressAnnotations) error {
	serversCfg, err := n.getBackendConfigure(ingress)
	if err != nil {
		return err
	}

	cfg := &configure{
		Cfg:         serversCfg,
		Annotations: ingress.ParsedAnnotations,
		TmplName:    config.ServerTmpl,
		MainTmpl:    config.MainServerTmpl,
		ConfName:    filepath.Join(config.ConfDir, n.ingress.Name+"-"+n.ingress.Namespace),
	}

	if err := n.generateConfigureBytes(cfg); err != nil {
		return err
	}

	klog.Infof("update %s-%s.conf successfully", n.ingress.Name, n.ingress.Namespace)

	if err := nginx.Reload(cfg.ConfName); err != nil {
		return err
	}

	return nil
}

func (n *NginxController) generateDefaultBackendTemplate(ingress annotations.IngressAnnotations) error {
	defaultCfg, err := n.getDefaultBackendConfigure(ingress)
	if err != nil {
		return err
	}

	conf := strings.Split(config.MainConf, ".")

	cfg := &configure{
		Cfg:         defaultCfg,
		Annotations: ingress.ParsedAnnotations,
		TmplName:    config.DefaultTmpl,
		MainTmpl:    config.NginxTmpl,
		ConfName:    conf[0],
	}

	if err := n.generateConfigureBytes(cfg); err != nil {
		return err
	}

	klog.Info(fmt.Sprintf("update %s successfully", filepath.Base(config.MainConf)))

	if err := nginx.Reload(cfg.ConfName); err != nil {
		return err
	}

	return nil
}

func (n *NginxController) getDefaultBackendConfigure(ingress annotations.IngressAnnotations) (*ingressv1.Configuration, error) {
	var servers []*ingressv1.Server
	var backends []*ingressv1.Backend

	svc, err := n.rr.GetService(n.ingress.Spec.DefaultBackend.Service.Name)
	if err != nil {
		return nil, err
	}

	backendPort := n.rr.GetSvcPort(*n.ingress.Spec.DefaultBackend)
	if backendPort == 0 {
		klog.ErrorS(fmt.Errorf("%s svc port not exists", svc.Name), "")
		return nil, fmt.Errorf("%s svc port not exists", svc.Name)
	}

	b := &ingressv1.Backend{
		Name:           svc.Name,
		NameSpace:      svc.Namespace,
		Port:           backendPort,
		Path:           "/",
		Annotations:    ingress.ParsedAnnotations,
		ServiceBackend: n.ingress.Spec.DefaultBackend.Service,
	}
	backends = append(backends, b)

	s := &ingressv1.Server{
		Name:      n.ingress.Name,
		NameSpace: n.ingress.Namespace,
		HostName:  "default",
		Paths:     backends,
	}

	servers = append(servers, s)

	return &ingressv1.Configuration{Servers: servers}, nil
}

func (n *NginxController) getBackendConfigure(ingress annotations.IngressAnnotations) (*ingressv1.Configuration, error) {
	var rule = n.ingress.Spec.Rules
	var servers = make([]*ingressv1.Server, len(rule))

	tls, err := n.generateTlsFile()
	if err != nil {
		klog.Warningf(fmt.Sprintf("failed to generate certificate and will not be able to use https"))
	}

	for k, v := range rule {
		var backend = make([]*ingressv1.Backend, len(v.HTTP.Paths))
		for bk, p := range v.HTTP.Paths {
			if err := n.checkIngressContent(&p, ingress.ParsedAnnotations); err != nil {
				return nil, err
			}

			svc, err := n.rr.GetService(p.Backend.Service.Name)
			if err != nil {
				return nil, err
			}

			backendPort := n.rr.GetSvcPort(p.Backend)
			if backendPort == 0 {
				klog.ErrorS(fmt.Errorf("svc port not exists"), fmt.Sprintf("svc port : %s not exists in namespace: %s", p.Backend.Service.Name, n.ingress.Namespace))
				return nil, fmt.Errorf("svc port not exists")
			}

			b := &ingressv1.Backend{
				IngName:        n.ingress.Name,
				Name:           svc.Name,
				NameSpace:      svc.Namespace,
				Path:           n.formatPath(p.Path, ingress),
				TargetPath:     p.Path,
				Port:           backendPort,
				ServiceBackend: p.Backend.Service,
				Annotations:    ingress.ParsedAnnotations,
			}
			backend = append(backend[:bk], b)
		}

		s := &ingressv1.Server{
			Name:      n.ingress.Name,
			NameSpace: n.ingress.Namespace,
			HostName:  v.Host,
			Paths:     backend,
			Tls:       tls[v.Host],
		}

		servers = append(servers[:k], s)
	}

	return &ingressv1.Configuration{Servers: servers}, nil
}

func (n *NginxController) generateTlsFile() (map[string]ingressv1.SSLCert, error) {
	if len(n.ingress.Spec.TLS) > 0 {
		return n.generateCaTlsFile()
	}

	return n.generateCrdTlsFile()
}

// Use Kubernetes internal self signed certificates
func (n *NginxController) generateCrdTlsFile() (map[string]ingressv1.SSLCert, error) {
	var ssl = ingressv1.SSLCert{}
	var ht = make(map[string]ingressv1.SSLCert)

	key := types.NamespacedName{Name: n.ingress.Name + "-secret", Namespace: n.ingress.Namespace}
	data, err := n.rr.GetTlsData(key)
	if err != nil {
		return ht, err
	}

	tlsPrefix := n.ingress.Name + "-" + n.ingress.Namespace + "-"

	for k, v := range data {
		file := filepath.Join(config.SslPath, tlsPrefix+k)
		if err := os.WriteFile(file, v, 0644); err != nil {
			return ht, err
		}
		if k == config.TlsCrt {
			ssl.TlsCrt = file
		} else if k == config.TlsKey {
			ssl.TlsKey = file
		}
	}

	for _, v := range n.ingress.Spec.Rules {
		ssl.TlsNoPass = true
		ht[v.Host] = ssl
	}

	return ht, nil
}

// Certificate signed with CA
func (n *NginxController) generateCaTlsFile() (map[string]ingressv1.SSLCert, error) {
	var ssl = ingressv1.SSLCert{}
	var ht = make(map[string]ingressv1.SSLCert)

	for _, secret := range n.ingress.Spec.TLS {
		for _, host := range secret.Hosts {
			key := types.NamespacedName{Name: secret.SecretName, Namespace: n.ingress.Namespace}
			data, err := n.rr.GetTlsData(key)
			if err != nil {
				return ht, err
			}
			for k, v := range data {
				hf := parser.GetDnsRegex(host)
				if hf == "" {
					return ht, fmt.Errorf("%s not a valid host", host)
				}
				file := filepath.Join(config.SslPath, host+"-"+n.ingress.Namespace+"-"+k)
				if err := os.WriteFile(file, v, 0644); err != nil {
					return ht, err
				}
				if k == config.TlsCrt {
					ssl.TlsCrt = file
				} else if k == config.TlsKey {
					ssl.TlsKey = file
				}
			}
			ssl.TlsNoPass = true
			ht[host] = ssl
		}
	}

	return ht, nil
}

func (n *NginxController) formatPath(path string, ingress annotations.IngressAnnotations) string {
	if ingress.ParsedAnnotations.Rewrite.EnableRegex || ingress.ParsedAnnotations.Rewrite.RewriteTarget != "" {
		path = "~ ^" + path
	}

	return path
}

func (n *NginxController) checkIngressContent(path *netv1.HTTPIngressPath, annotations *annotations.Ingress) error {
	var err error
	var info string
	if annotations.Rewrite.EnableRegex && *path.PathType != "ImplementationSpecific" {
		klog.Warningf(
			fmt.Sprintf("the value of pathType should be define as ImplementationSpecific in ingress: %s, namespace: %s", n.ingress.Name, n.ingress.Namespace))
	}

	if parser.IsRegexPatternRegex(path.Path) && !annotations.Rewrite.EnableRegex && annotations.Rewrite.RewriteTarget == "" {
		err = fmt.Errorf("the value of ingress path: %s looks like regexp. please add corresponding annotations such as: rewrite-target or enable-regex", path.Path)
		info = fmt.Sprintf("missing valid annotations")
		klog.ErrorS(err, info)
		return err
	}

	if annotations.Rewrite.EnableRegex && !parser.IsRegexPatternRegex(path.Path) {
		err = fmt.Errorf("the path value: %s in ingress should be a valid regexp because enable-regex is used in annotations", path.Path)
		info = fmt.Sprintf("the path of ingress is not a valid regular expression")
		klog.ErrorS(err, info)
		return err
	}

	if annotations.Rewrite.RewriteTarget != "" && !parser.IsRegexPatternRegex(path.Path) {
		err = fmt.Errorf("the path value: %s in ingress should be a valid regexp because rewrite-target is used in annotations", path.Path)
		info = fmt.Sprintf("the path of ingress is not a valid regular expression")
		klog.ErrorS(err, info)
		return err
	}

	return nil
}
