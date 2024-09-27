package redirect

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	serverRedirectHostAnnotation = "redirect-host"
	serverRedirectPathAnnotation = "redirect-path"
)

var redirectAnnotation = parser.Annotation{
	Group: "redirect",
	Annotations: parser.AnnotationFields{
		serverRedirectHostAnnotation: {
			Doc: "return 301 e.g: `http://*`, required",
		},
		serverRedirectPathAnnotation: {
			Doc: "match path redirect, e.g: `/aaa`, optional",
		},
	},
}

type Config struct {
	Host string `json:"host"`
	Path string `json:"path"`
}

type redirect struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &redirect{}
}

func (r *redirect) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.Host, err = parser.GetStringAnnotation(serverRedirectHostAnnotation, ing, redirectAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", serverRedirectHostAnnotation)
		}
	}

	config.Path, err = parser.GetStringAnnotation(serverRedirectPathAnnotation, ing, redirectAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", serverRedirectPathAnnotation)
		}
	}

	if !r.check(config) {
		return nil, errors.NewInvalidAnnotationsContentError(serverRedirectHostAnnotation, config.Host)
	}

	return config, nil
}

func (r *redirect) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, redirectAnnotation.Annotations)
}

func (r *redirect) check(cfg *Config) bool {
	if cfg.Path != "" && !parser.IsValidHost(cfg.Host) {
		return false
	}

	return true
}
