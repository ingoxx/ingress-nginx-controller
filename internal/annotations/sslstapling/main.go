package sslstapling

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
)

const (
	sslStaplingVerify = "ssl-stapling-verify"
	sslStapling       = "ssl-stapling"
)

type SSl struct {
	r resolver.Resolver
}

type Config struct {
	SSllStaplingVerify bool `json:"ssl-stapling-verify"`
	SSlStapling        bool `json:"sslstapling-stapling"`
}

var sslAnnotations = parser.Annotation{
	Group: "sslStapling",
	Annotations: parser.AnnotationFields{
		sslStaplingVerify: {
			Doc: "switch ssl stapling verify",
		},
		sslStapling: {
			Doc: "switch ssl stapling",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &SSl{}
}

func (p *SSl) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.SSllStaplingVerify, err = parser.GetBoolAnnotations(sslStaplingVerify, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", sslStaplingVerify)
		}
		config.SSllStaplingVerify = false
	}

	config.SSlStapling, err = parser.GetBoolAnnotations(sslStapling, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", sslStapling)
		}
		config.SSlStapling = false
	}

	return config, nil
}

func (p *SSl) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, sslAnnotations.Annotations)
}
