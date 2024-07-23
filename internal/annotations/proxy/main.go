package proxy

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
)

const (
	proxyPathAnnotation          = "proxy-path"
	proxyHostAnnotation          = "proxy-host"
	proxyTargetAnnotation        = "proxy-target"
	proxySSLAnnotation           = "proxy-ssl"
	proxyEnableRewriteAnnotation = "proxy-enable-rewrite"
)

type Proxy struct {
	r resolver.Resolver
}

type Config struct {
	ProxyPath          string `json:"proxy-path"`
	ProxyHost          string `json:"proxy-host"`
	ProxyTarget        string `json:"proxy-target"`
	ProxySSL           bool   `json:"proxy-ssl"`
	ProxyEnableRewrite bool   `json:"proxy-enable-rewrite"`
}

var proxyAnnotation = parser.Annotation{
	Group: "proxy",
	Annotations: parser.AnnotationFields{
		proxyPathAnnotation: {
			Doc: "matching target path, e.g: /aaa/bbb or /aaa/bbb(/|$)(.*), required",
		},
		proxyHostAnnotation: {
			Doc: "url link outside the cluster, e.g: http://ccc.com, required",
		},
		proxySSLAnnotation: {
			Doc: "choose https, e.g: https://ccc.com, optional",
		},
		proxyTargetAnnotation: {
			Doc: "when the ProxyEnableRewrite is true, choose it, e.g: /$1,$2..., optional",
		},
		proxyEnableRewriteAnnotation: {
			Doc: "whether to enable the rewrite function, optional",
		},
	},
}

func NewParser(r resolver.Resolver) *Proxy {
	return &Proxy{}
}

func (p *Proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.ProxyPath, err = parser.GetStringAnnotation(proxyPathAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.ProxyHost, err = parser.GetStringAnnotation(proxyHostAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.ProxyTarget, err = parser.GetStringAnnotation(proxyTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.ProxySSL, err = parser.GetBoolAnnotations(proxySSLAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.ProxyEnableRewrite, err = parser.GetBoolAnnotations(proxyEnableRewriteAnnotation, ing)
	if err != nil {
		return nil, err
	}

	if config.ProxyPath == "" || config.ProxyHost == "" {
		return config, fmt.Errorf("the %s and %s cannot be empty when using the proxy function", proxyPathAnnotation, proxyHostAnnotation)
	}

	if config.ProxyEnableRewrite {
		if config.ProxyTarget == "" {
			return config, fmt.Errorf("if %s is true, the %s value cannot be empty in annotations", proxyEnableRewriteAnnotation, proxyTargetAnnotation)
		}
	}

	return config, nil
}

func (p *Proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}
