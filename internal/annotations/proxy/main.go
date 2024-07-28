package proxy

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	proxyPathAnnotation          = "proxy-path"
	proxyHostAnnotation          = "proxy-host"
	proxyTargetAnnotation        = "proxy-target"
	proxySSLAnnotation           = "proxy-sslstapling"
	proxyEnableRewriteAnnotation = "proxy-enable-rewrite"
	proxyEnableRegex             = "proxy-enable-regex"
)

// The role of a proxy is to forward traffic requests to the cluster to the outside of the cluster
type proxy struct {
	r resolver.Resolver
}

type Config struct {
	ProxyPath          string `json:"proxy-path"`
	ProxyHost          string `json:"proxy-host"`
	ProxyTarget        string `json:"proxy-target"`
	ProxySSL           bool   `json:"proxy-sslstapling"`
	ProxyEnableRewrite bool   `json:"proxy-enable-rewrite"`
	ProxyEnableRegex   bool   `json:"proxy-enable-regex"`
}

var proxyAnnotation = parser.Annotation{
	Group: "proxy",
	Annotations: parser.AnnotationFields{
		proxyPathAnnotation: {
			Doc: "matching target path, e.g: /aaa/bbb or regex: /aaa/bbb(/|$)(.*), required",
		},
		proxyHostAnnotation: {
			Doc: "url link outside the cluster, e.g: ccc.com or 1.1.1.1, required",
		},
		proxySSLAnnotation: {
			Doc: "if true, the proxy_pass will be https, optional",
		},
		proxyTargetAnnotation: {
			Doc: "when the ProxyEnableRewrite is true, choose it, e.g: /$1,$2..., optional",
		},
		proxyEnableRewriteAnnotation: {
			Doc: "whether to enable the rewrite function, optional",
		},
		proxyEnableRegex: {
			Doc: "This annotation defines if the paths defined on an Ingress use regular expressions. To use regex on path\n\t\t\tthe pathType should also be defined as 'ImplementationSpecific'., optional",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &proxy{}
}

func (p *proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.ProxyPath, err = parser.GetStringAnnotation(proxyPathAnnotation, ing)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("%s not allow empty", proxyPathAnnotation))
		return config, err
	}

	config.ProxyHost, err = parser.GetStringAnnotation(proxyHostAnnotation, ing)
	if err != nil {
		klog.ErrorS(err, fmt.Sprintf("%s not allow empty", proxyHostAnnotation))
		return config, err
	}

	config.ProxyTarget, err = parser.GetStringAnnotation(proxyTargetAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", proxyTargetAnnotation)
		}
	}

	config.ProxySSL, err = parser.GetBoolAnnotations(proxySSLAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", proxySSLAnnotation)
		}
	}

	config.ProxyEnableRewrite, err = parser.GetBoolAnnotations(proxyEnableRewriteAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", proxyEnableRewriteAnnotation)
		}
	}

	config.ProxyEnableRegex, err = parser.GetBoolAnnotations(proxyEnableRegex, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", proxyEnableRegex)
		}
	}

	if config.ProxyTarget != "" && !config.ProxyEnableRegex {
		return config, fmt.Errorf("if annotations %s is not empty, annotations %s must be true", proxyTargetAnnotation, proxyEnableRegex)
	}

	if parser.PassIsIp(config.ProxyHost) && config.ProxySSL {
		return config, fmt.Errorf("if %s is true, the %s not allow ip", proxySSLAnnotation, proxyHostAnnotation)
	}

	return config, nil
}

func (p *proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}
