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
	proxyPathAnnotation   = "proxy-path"
	proxyHostAnnotation   = "proxy-host"
	proxyTargetAnnotation = "proxy-target"
	proxySSLAnnotation    = "proxy-sslstapling"
	proxyEnableRegex      = "proxy-enable-regex"
)

// The role of a proxy is to forward traffic requests to the cluster to the outside of the cluster
type proxy struct {
	r resolver.Resolver
}

type Config struct {
	ProxyPath        string `json:"proxy-path"`
	ProxyHost        string `json:"proxy-host"`
	ProxyTarget      string `json:"proxy-target"`
	ProxyTargetPath  string `json:"proxy-target-path"`
	ProxySSL         bool   `json:"proxy-sslstapling"`
	ProxyEnableRegex bool   `json:"proxy-enable-regex"`
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
	config.ProxyPath, err = parser.GetStringAnnotation(proxyPathAnnotation, ing, proxyAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", proxyPathAnnotation)
		}
	}

	config.ProxyHost, err = parser.GetStringAnnotation(proxyHostAnnotation, ing, proxyAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", proxyHostAnnotation)
		}
	}

	config.ProxyTarget, err = parser.GetStringAnnotation(proxyTargetAnnotation, ing, proxyAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", proxyTargetAnnotation)
		}
	}

	config.ProxySSL, err = parser.GetBoolAnnotations(proxySSLAnnotation, ing, proxyAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", proxySSLAnnotation)
		}
	}

	config.ProxyEnableRegex, err = parser.GetBoolAnnotations(proxyEnableRegex, ing, proxyAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", proxyEnableRegex)
		}
	}

	//check
	if err := p.check(config); err != nil {
		return nil, err
	}

	config.ProxyTargetPath = config.ProxyPath
	config.ProxyPath = p.formatProxyPath(config.ProxyPath, *config)

	return config, nil
}

func (p *proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}

func (p *proxy) formatProxyPath(path string, proxyConfig Config) string {
	if proxyConfig.ProxyTarget != "" || proxyConfig.ProxyEnableRegex {
		path = "~ ^" + proxyConfig.ProxyPath
	}

	return path
}

func (p *proxy) check(config *Config) error {
	var err error
	if config.ProxyTarget != "" && !parser.IsTargetPathRegex(config.ProxyTarget) {
		klog.ErrorS(errors.NewValidationError(proxyTargetAnnotation),
			proxyAnnotation.Annotations[proxyTargetAnnotation].Doc)
		return errors.NewValidationError(proxyTargetAnnotation)
	}

	if parser.IsRegexPatternRegex(config.ProxyPath) && !config.ProxyEnableRegex && config.ProxyTarget == "" {
		err = errors.NewValidationError(proxyPathAnnotation)
		klog.ErrorS(err,
			fmt.Sprintf("the value of annotations %s: %s looks like regexp. please add corresponding annotations such as: %s or %s",
				proxyPathAnnotation, config.ProxyPath, proxyTargetAnnotation, proxyEnableRegex),
		)
		return err
	}

	if config.ProxyEnableRegex && !parser.IsRegexPatternRegex(config.ProxyPath) {
		err = errors.NewValidationError(proxyPathAnnotation)
		klog.ErrorS(err,
			fmt.Sprintf("the %s value in annotations should be a valid regexp because %s is used in annotations", proxyPathAnnotation, proxyEnableRegex))
		return err
	}

	if config.ProxyTarget != "" && !parser.IsRegexPatternRegex(config.ProxyPath) {
		err = errors.NewValidationError(proxyPathAnnotation)
		klog.ErrorS(err,
			fmt.Sprintf("the %s value in annotations should be a valid regexp because %s is used in annotations", proxyPathAnnotation, proxyTargetAnnotation))
		return err
	}

	if parser.PassIsIp(config.ProxyHost) && config.ProxySSL {
		err = errors.NewValidationError(proxyHostAnnotation)
		klog.ErrorS(err, proxyAnnotation.Annotations[proxyTargetAnnotation].Doc)
		return err
	}

	return nil
}
