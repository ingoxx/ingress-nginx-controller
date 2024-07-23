package rewrite

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"strings"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteSSLAnnotation         = "rewrite-ssl"
	rewriteEnableRegexAnnotation = "rewrite-enable-regex"
	rewriteAllowListAnnotation   = "rewrite-ip-allow-list"
	rewriteDenyListAnnotation    = "rewrite-ip-deny-list"
)

type Proxy struct {
	r resolver.Resolver
}

type Config struct {
	RewriteTarget      string   `json:"rewrite-target"`
	RewriteSSL         bool     `json:"rewrite-ssl"`
	RewriteEnableRegex bool     `json:"rewrite-enable-regex"`
	RewriteIpAllowList []string `json:"rewrite-ip-allow-list"`
	RewriteIpDenyList  []string `json:"rewrite-ip-deny-list"`
}

var proxyAnnotation = parser.Annotation{
	Group: "rewrite",
	Annotations: parser.AnnotationFields{
		rewriteTargetAnnotation: {
			Doc: "",
		},
		rewriteSSLAnnotation: {
			Doc: "",
		},
		rewriteEnableRegexAnnotation: {
			Doc: "",
		},
		rewriteAllowListAnnotation: {
			Doc: "",
		},
		rewriteDenyListAnnotation: {
			Doc: "",
		},
	},
}

func NewParser(r resolver.Resolver) *Proxy {
	return &Proxy{}
}

func (p *Proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	rewriteIpAllowList, err := parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.RewriteIpAllowList = strings.Split(rewriteIpAllowList, ",")

	rewriteIpDenyList, err := parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.RewriteIpDenyList = strings.Split(rewriteIpDenyList, ",")

	config.RewriteSSL, err = parser.GetBoolAnnotations(rewriteSSLAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.RewriteEnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (p *Proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}
