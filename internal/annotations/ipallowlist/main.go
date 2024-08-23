package ipallowlist

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"strings"
)

const (
	allowListAnnotation = "allowList"
)

type ipallowList struct {
	r resolver.Resolver
}

type SourceRange struct {
	CIDR []string `json:"cidr,omitempty"`
}

var ipAllowListAnnotations = parser.Annotation{
	Group: "ipallowlist",
	Annotations: parser.AnnotationFields{
		allowListAnnotation: {
			Doc: "allow ip, eg: 1.1.1.1",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &ipallowList{}
}

// Parse Only valid for the backend within the current ingress rule
// e.g. 10.0.0.8/16,11.0.0.9/16
func (p *ipallowList) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	val, err := parser.GetStringAnnotation(allowListAnnotation, ing, ipAllowListAnnotations.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty slice", allowListAnnotation)
		}
	}

	aliases := sets.NewString()
	for _, alias := range strings.Split(val, ",") {
		alias = strings.TrimSpace(alias)
		if alias == "" {
			continue
		}
		if !parser.PassIsIp(alias) {
			return nil, fmt.Errorf("the annotation %s does not contain a valid IP address", allowListAnnotation)
		}

		if !aliases.Has(alias) {
			aliases.Insert(alias)
		}
	}

	l := aliases.List()

	return &SourceRange{l}, nil
}

func (p *ipallowList) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, ipAllowListAnnotations.Annotations)
}
