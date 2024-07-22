package rewrite

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteSSLAnnotation         = "rewrite-ssl"
	rewriteEnableRegexAnnotation = "rewrite-enable-regex"
)

type Proxy struct {
	r resolver.Resolver
}

type Config struct {
	Target      string `json:"target"`
	EnableSSL   bool   `json:"enable_ssl"`
	EnableRegex bool   `json:"enable_regex"`
}

var proxyAnnotation = parser.Annotation{
	Group: "rewrite",
	Annotations: parser.AnnotationFields{
		rewriteTargetAnnotation:      rewriteTargetAnnotation,
		rewriteSSLAnnotation:         rewriteSSLAnnotation,
		rewriteEnableRegexAnnotation: rewriteEnableRegexAnnotation,
	},
}

func NewParser(r resolver.Resolver) *Proxy {
	return &Proxy{}
}

func (p *Proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.Target, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.Target, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.EnableSSL, err = parser.GetBoolAnnotations(rewriteSSLAnnotation, ing)
	if err != nil {
		return nil, err
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (p *Proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, proxyAnnotation.Annotations)
}
