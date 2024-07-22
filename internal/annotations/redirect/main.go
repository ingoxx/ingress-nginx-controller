package redirect

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
)

const (
	serverRedirectAnnotation = "redirect"
)

var redirectAnnotation = parser.Annotation{
	Group: "redirect",
	Annotations: parser.AnnotationFields{
		serverRedirectAnnotation: serverRedirectAnnotation,
	},
}

type Config struct {
	Host string `json:"host"`
}

type Redirect struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) *Redirect {
	return &Redirect{}
}

func (r *Redirect) Parse(ing *ingressv1.Ingress) (interface{}, error) {

	return nil, nil
}

func (r *Redirect) Validate(anns map[string]string) error {
	return nil
}
