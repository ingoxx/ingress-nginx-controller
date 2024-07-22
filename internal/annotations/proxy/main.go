package proxy

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
)

const (
	proxyPathAnnotation   = "proxy-path"
	proxyHostAnnotation   = "proxy-host"
	proxyTargetAnnotation = "proxy-target"

	proxySSLAnnotation         = "proxy-ssl"
	proxyEnableRegexAnnotation = "proxy-enable-regex"
)

type Proxy struct {
	r resolver.Resolver
}

type Config struct {
	ProxyPath        string `json:"proxy-path"`
	ProxyHost        string `json:"proxy-host"`
	ProxyTarget      string `json:"proxy-target"`
	ProxySSL         bool   `json:"proxy-ssl"`
	ProxyEnableRegex bool   `json:"proxy-enable-regex"`
}

var proxyAnnotation = parser.Annotation{
	Group: "proxy",
	Annotations: parser.AnnotationFields{
		proxyPathAnnotation:        proxyPathAnnotation,
		proxyHostAnnotation:        proxyHostAnnotation,
		proxySSLAnnotation:         proxySSLAnnotation,
		proxyEnableRegexAnnotation: proxyEnableRegexAnnotation,
		proxyTargetAnnotation:      proxyTargetAnnotation,
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

	config.ProxyEnableRegex, err = parser.GetBoolAnnotations(proxyEnableRegexAnnotation, ing)
	if err != nil {
		return nil, err
	}

	if config.ProxyPath == "" || config.ProxyHost == "" {
		return config, fmt.Errorf("the proxy function is to match the specified path and forward it to the external backend, so the %s or %s values of annotations cannot be empty", proxyPathAnnotation, proxyHostAnnotation)
	}

	return config, nil
}

func (p *Proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}
