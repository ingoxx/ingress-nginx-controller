package ipdenylist

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"strings"
)

const (
	denyListAnnotation = "denyList"
)

type ipdenyList struct {
	r resolver.Resolver
}

type SourceRange struct {
	CIDR []string `json:"cidr,omitempty"`
}

var ipDenyListAnnotations = parser.Annotation{
	Group: "ipdenylist",
	Annotations: parser.AnnotationFields{
		denyListAnnotation: {
			Doc: "deny ip, e.g: `2.2.2.2`, required",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &ipdenyList{}
}

func (p *ipdenyList) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	val, err := parser.GetStringAnnotation(denyListAnnotation, ing, ipDenyListAnnotations.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty slice", denyListAnnotation)
		}
	}

	aliases := sets.NewString()
	for _, alias := range strings.Split(val, ",") {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		if !parser.PassIsIp(alias) {
			return nil, errors.NewInvalidAnnotationsContentError(denyListAnnotation, alias)
		}

		if !aliases.Has(alias) {
			aliases.Insert(alias)
		}
	}

	l := aliases.List()

	return &SourceRange{l}, nil
}

func (p *ipdenyList) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, ipDenyListAnnotations.Annotations)
}
