package ipallowlist

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"k8s.io/apimachinery/pkg/util/sets"
	"strings"
)

const (
	allowListAnnotation = "allowList"
)

type Proxy struct {
	r resolver.Resolver
}

type SourceRange struct {
	CIDR []string `json:"cidr,omitempty"`
}

var ipAllowListAnnotations = parser.Annotation{
	Group: "ipallowlist",
	Annotations: parser.AnnotationFields{
		allowListAnnotation: {
			Doc: "",
		},
	},
}

func NewParser(r resolver.Resolver) *Proxy {
	return &Proxy{}
}

// Parse Only valid for the backend within the current ingress rule
// e.g. 10.0.0.8/16,11.0.0.9/16
func (p *Proxy) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	val, err := parser.GetStringAnnotation(allowListAnnotation, ing)
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
	return parser.CheckAnnotations(anns, ipAllowListAnnotations.Annotations)
}
