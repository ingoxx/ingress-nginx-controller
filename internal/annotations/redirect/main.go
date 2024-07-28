package redirect

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	serverRedirectAnnotation = "redirect"
)

var redirectAnnotation = parser.Annotation{
	Group: "redirect",
	Annotations: parser.AnnotationFields{
		serverRedirectAnnotation: {
			Doc: "return 301 http://*",
		},
	},
}

type Config struct {
	Host string `json:"host"`
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
	config.Host, err = parser.GetStringAnnotation(serverRedirectAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", serverRedirectAnnotation)
		}
	}

	return config, nil
}

func (r *redirect) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, redirectAnnotation.Annotations)
}
