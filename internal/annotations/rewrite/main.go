package rewrite

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteEnableRegexAnnotation = "enable-regex"
)

type rewrite struct {
	r resolver.Resolver
}

type Config struct {
	RewriteTarget string `json:"rewrite-target"`
	EnableRegex   bool   `json:"enable-regex"`
}

var rewriteAnnotation = parser.Annotation{
	Group: "rewrite",
	Annotations: parser.AnnotationFields{
		rewriteTargetAnnotation: {
			Doc: "It can contain regular characters and captured \n\t\t\tgroups specified as '$1', '$2', '/asd'.",
		},
		rewriteEnableRegexAnnotation: {
			Doc: "This annotation defines if the paths defined on an Ingress use regular expressions. To use regex on path\n\t\t\tthe pathType should also be defined as 'ImplementationSpecific'.",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &rewrite{}
}

func (p *rewrite) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing, rewriteAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", rewriteTargetAnnotation)
		}
	}

	config.EnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing, rewriteAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", rewriteEnableRegexAnnotation)
		}
	}

	if config.RewriteTarget != "" && !parser.IsTargetPathRegex(config.RewriteTarget) {
		klog.ErrorS(errors.NewValidationError(rewriteTargetAnnotation),
			rewriteAnnotation.Annotations[rewriteTargetAnnotation].Doc)
		return config, errors.NewValidationError(rewriteTargetAnnotation)
	}

	return config, nil
}

func (p *rewrite) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, rewriteAnnotation.Annotations)
}
