package controller

import (
	"bytes"
	"context"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/ipallowlist"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/proxy"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/nginx"
	"k8s.io/klog/v2"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"
	"text/template"
)

const (
	nginxTmpl          = "/workspace/rootfs/etc/nginx/template/nginx.tmpl"
	proxyTmpl          = "/workspace/rootfs/etc/nginx/template/proxy.tmpl"
	sslTmpl            = "/workspace/rootfs/etc/nginx/template/ssl.tmpl"
	serverTmpl         = "/workspace/rootfs/etc/nginx/template/server.tmpl"
	redirectTmpl       = "/workspace/rootfs/etc/nginx/template/redirect.tmpl"
	defaultBackendTmpl = "/workspace/rootfs/etc/nginx/template/defaultBackend.tmpl"
	ipAllowListTmpl    = "/workspace/rootfs/etc/nginx/template/ipAllowList.tmpl"
	ipDenyListTmpl     = "/workspace/rootfs/etc/nginx/template/ipDenyList.tmpl"
	nginxConf          = "/etc/nginx/nginx.conf"
	sslPath            = "/etc/nginx/ssl"
	testConf           = "/etc/nginx/nginx.conf.test"
)

type renderNginxConfigure struct {
	HostName  string                  `json:"host_name"`
	Path      string                  `json:"path"`
	RenderSsl bool                    `json:"render_ssl"`
	AllowList ipallowlist.SourceRange `json:"allow_list"`
	DenyList  ipallowlist.SourceRange `json:"deny_list"`
	Proxy     proxy.Config            `json:"proxy"`
	Redirect  []string                `json:"redirect"`
}

type configure struct {
	Server      *ingressv1.Server
	Annotations *annotations.Ingress
	ServerTpl   bytes.Buffer
}

type NginxConfigure struct {
	client client.Client
	ctx    context.Context
	rr     resolver.Resolver
}

func NewNginxConfigure(r client.Client, ctx context.Context, rr resolver.Resolver) *NginxConfigure {
	n := &NginxConfigure{
		client: r,
		ctx:    ctx,
		rr:     rr,
	}
	return n
}

func (n *NginxConfigure) generateBackend(backends []*ingressv1.Backend) (bytes.Buffer, error) {
	var proxyBytes bytes.Buffer
	var backendPort int32

	proxyTmplStr, err := os.ReadFile(proxyTmpl)
	if err != nil {
		return proxyBytes, err
	}

	for _, v := range backends {
		svc, err := n.rr.GetService(v.ServiceBackend.Name)
		if err != nil {
			return proxyBytes, err
		}
		for _, p := range svc.Spec.Ports {
			if p.Name == v.ServiceBackend.Name || p.Port == v.ServiceBackend.Port.Number {
				backendPort = p.Port
				break
			}
		}
		var data = renderNginxConfigure{
			HostName: v.Name + strconv.Itoa(int(backendPort)),
			Path:     v.Path,
		}
		parse, err := template.New("proxyTmpl").Parse(string(proxyTmplStr))
		if err != nil {
			return proxyBytes, err
		}

		if err := parse.Execute(&proxyBytes, data); err != nil {
			return proxyBytes, err
		}

	}

	return proxyBytes, nil
}

func (n *NginxConfigure) generateAllowList() (*template.Template, error) {
	return nil, nil
}

func (n *NginxConfigure) generateDenyList() (*template.Template, error) {
	return nil, nil
}

func (n *NginxConfigure) generateProxy(config *configure) (bytes.Buffer, error) {
	var proxyBytes bytes.Buffer
	file, err := os.ReadFile(proxyTmpl)
	if err != nil {
		return proxyBytes, err
	}

	return proxyBytes, nil
}

func (n *NginxConfigure) generateRedirect(config *configure) (bytes.Buffer, error) {
	var redirectBytes bytes.Buffer
	file, err := os.ReadFile(redirectTmpl)
	if err != nil {
		return redirectBytes, err
	}

	redirect := n.readyRenderData(config)
	for _, v := range redirect.Redirect {
		var data = struct {
			HostName string
			Path     string
		}{
			Path:     strings.Split(v, ",")[0],
			HostName: strings.Split(v, ",")[1],
		}

		proxyTpl, err := template.New("redirect").Parse(string(file))
		if err != nil {
			return redirectBytes, err
		}

		if err := proxyTpl.Execute(&redirectBytes, data); err != nil {
			return redirectBytes, err
		}

	}
	return redirectBytes, nil
}

func (n *NginxConfigure) readyRenderData(config *configure) renderNginxConfigure {
	var data renderNginxConfigure
	if len(config.Annotations.AllowList.CIDR) > 0 {
		data.AllowList = config.Annotations.AllowList.CIDR
	}

	if len(config.Annotations.DenyList.CIDR) > 0 {
		data.DenyList = config.Annotations.DenyList.CIDR
	}

	if len(config.Annotations.Redirect) > 0 {
		data.Redirect = config.Annotations.Redirect
	}

	if len(config.Annotations.Proxy) > 0 {
		data.Proxy = config.Annotations.Proxy
	}

	data.HostName = config.Server.HostName

	return data
}

func (n *NginxConfigure) generateServer(config *configure) error {
	serverStr, err := os.ReadFile(serverTmpl)
	if err != nil {
		return err
	}

	serverTemp, err := template.New("serverMain").Parse(string(serverStr))
	if err != nil {
		log.Fatal("Error parsing nginx.conf.tmpl:", err)
	}

	var serverByte bytes.Buffer
	var backendBytes bytes.Buffer

	data := n.readyRenderData(config)

	proxyBackendBytes, err := n.generateProxy(config)
	if err != nil {
		return err
	}

	_, err = serverTemp.New("proxy").Parse(proxyBackendBytes.String())
	if err != nil {
		return err
	}

	redirectBackendBytes, err := n.generateRedirect(config)
	if err != nil {
		return err
	}

	if config.Server != nil {
		_, err = serverTemp.New("redirect").Parse(redirectBackendBytes.String())
		if err != nil {
			return err
		}

		backendBytes, err = n.generateBackend(config.Server.Paths)
		if err != nil {
			return err
		}
	}

	_, err = serverTemp.New("backend").Parse(backendBytes.String())
	if err != nil {
		return err
	}

	if err := serverTemp.Execute(&serverByte, data); err != nil {
		return err
	}

	return nil
}

func (n *NginxConfigure) generateNginxConfigure(cfg *ingressv1.Configuration, annotations *annotations.Ingress) error {
	mainTmplStr, err := os.ReadFile(nginxTmpl)
	if err != nil {
		return err
	}

	mainTmpl, mainErr := template.New("main").Parse(string(mainTmplStr))
	if mainErr != nil {
		return mainErr
	}

	var config = new(configure)
	var dynamicTmpl bytes.Buffer
	config.ServerTpl = dynamicTmpl
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

	if err := n.generateServer(config); err != nil {
		return err
	}

	//生成的server块: server {}插入到指定位置
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
	if err := os.WriteFile(testConf, b, 0644); err != nil {
		klog.ErrorS(err, "an error occurred while writing the generated template content to testConf.")
		return err
	}

	return nil
}

func (n *NginxConfigure) GenerateNginxConfigure(ingress annotations.IngressAnnotations) error {
	var cfg *ingressv1.Configuration
	if len(ingress.Ingress.Spec.Rules) > 0 {
		cfg = n.getBackendConfigure(ingress)
	} else {
		cfg = n.getDefaultBackendConfigure(ingress)
	}

	if cfg == nil {
		return fmt.Errorf("fail to get server list in ingress: %s", ingress.Ingress.Name)
	}

	if err := n.generateNginxConfigure(cfg, ingress.ParsedAnnotations); err != nil {
		return err
	}

	if err := nginx.Reload(nginxConf, testConf); err != nil {
		return err
	}

	return nil
}

func (n *NginxConfigure) getDefaultBackendConfigure(ingress annotations.IngressAnnotations) *ingressv1.Configuration {
	var servers = make([]*ingressv1.Server, 1)
	var backend = make([]*ingressv1.Backend, 1)
	b := &ingressv1.Backend{
		Path:           "/",
		ServiceBackend: ingress.Ingress.Spec.DefaultBackend.Service,
	}

	backend = append(backend, b)

	s := &ingressv1.Server{
		Name:      ingress.Ingress.Name,
		NameSpace: ingress.Ingress.Namespace,
		HostName:  "_",
		Paths:     backend,
	}
	servers = append(servers, s)

	return &ingressv1.Configuration{Servers: servers}
}

func (n *NginxConfigure) getBackendConfigure(ingress annotations.IngressAnnotations) *ingressv1.Configuration {
	var rule = ingress.Ingress.Spec.Rules
	var servers = make([]*ingressv1.Server, len(rule))

	for _, v := range rule {
		var backend = make([]*ingressv1.Backend, len(v.HTTP.Paths))
		for _, p := range v.HTTP.Paths {
			b := &ingressv1.Backend{
				Name:           v.Host,
				Path:           p.Path,
				ServiceBackend: p.Backend.Service,
			}
			backend = append(backend, b)
		}

		s := &ingressv1.Server{
			Name:      ingress.Ingress.Name,
			NameSpace: ingress.Ingress.Namespace,
			HostName:  v.Host,
			Paths:     backend,
		}

		servers = append(servers, s)
	}
	return &ingressv1.Configuration{Servers: servers}
}
