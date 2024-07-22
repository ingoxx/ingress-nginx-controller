package ipdenylist

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

const (
	denyListAnnotation = "denyList"
)

type Proxy struct {
	r resolver.Resolver
}

type SourceRange struct {
	CIDR []string `json:"cidr,omitempty"`
}

var ipDenyListAnnotations = parser.Annotation{
	Group: "ipdenylist",
	Annotations: parser.AnnotationFields{
		denyListAnnotation: denyListAnnotation,
	},
}

func NewParser(r resolver.Resolver) *Proxy {
	return &Proxy{}
}

func (p *Proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	val, err := parser.GetStringAnnotation(denyListAnnotation, ing)
	if err != nil {
		return nil, err
	}

	aliases := sets.NewString()
	for _, alias := range strings.Split(val, ",") {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}

		if !aliases.Has(alias) {
			aliases.Insert(alias)
		}
	}

	l := aliases.List()

	return &SourceRange{l}, nil
}

func (p *Proxy) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, ipDenyListAnnotations.Annotations)
}
