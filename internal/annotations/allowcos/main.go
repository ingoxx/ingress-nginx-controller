package allowcos

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	allowCosAnnotation = "use-cos"
)

var cosAnnotation = parser.Annotation{
	Group: "allowCos",
	Annotations: parser.AnnotationFields{
		allowCosAnnotation: {
			Doc: "allow cos, e.g: `true or false`, required",
		},
	},
}

type Config struct {
	AllowCos bool `json:"allow_cos"`
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
	config.AllowCos, err = parser.GetBoolAnnotations(allowCosAnnotation, ing, cosAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", allowCosAnnotation)
		}
	}

	return config, nil
}

func (r *redirect) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, cosAnnotation.Annotations)
}
