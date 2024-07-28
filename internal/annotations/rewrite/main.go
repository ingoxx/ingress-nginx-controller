package rewrite

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"k8s.io/klog/v2"
	"strings"
)

const (
	rewriteTargetAnnotation      = "rewrite-target"
	rewriteSSLAnnotation         = "rewrite-sslstapling"
	rewriteEnableRegexAnnotation = "rewrite-enable-regex"
	rewriteAllowListAnnotation   = "rewrite-ip-allow-list"
	rewriteDenyListAnnotation    = "rewrite-ip-deny-list"
)

type rewrite struct {
	r resolver.Resolver
}

type Config struct {
	RewriteTarget      string   `json:"rewrite-target"`
	RewriteEnableRegex bool     `json:"rewrite-enable-regex"`
	RewriteIpAllowList []string `json:"rewrite-ip-allow-list"`
	RewriteIpDenyList  []string `json:"rewrite-ip-deny-list"`
}

var rewriteAnnotation = parser.Annotation{
	Group: "rewrite",
	Annotations: parser.AnnotationFields{
		rewriteTargetAnnotation: {
			Doc: "rewrite target",
		},
		rewriteSSLAnnotation: {
			Doc: "sslstapling",
		},
		rewriteEnableRegexAnnotation: {
			Doc: "regex",
		},
		rewriteAllowListAnnotation: {
			Doc: "ip",
		},
		rewriteDenyListAnnotation: {
			Doc: "ip",
		},
	},
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &rewrite{}
}

func (p *rewrite) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	config := &Config{}
	config.RewriteTarget, err = parser.GetStringAnnotation(rewriteTargetAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", rewriteTargetAnnotation)
		}
	}

	rewriteIpAllowList, err := parser.GetStringAnnotation(rewriteAllowListAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", rewriteTargetAnnotation)
		}
	}

	config.RewriteIpAllowList = strings.Split(rewriteIpAllowList, ",")
	rewriteIpDenyList, err := parser.GetStringAnnotation(rewriteDenyListAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", rewriteDenyListAnnotation)
		}
	}

	config.RewriteIpDenyList = strings.Split(rewriteIpDenyList, ",")

	config.RewriteEnableRegex, err = parser.GetBoolAnnotations(rewriteEnableRegexAnnotation, ing)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", rewriteEnableRegexAnnotation)
		}
	}

	if config.RewriteTarget != "" && !config.RewriteEnableRegex {
		return config, fmt.Errorf("if annotations %s is not empty, annotations %s must be true", rewriteTargetAnnotation, rewriteEnableRegexAnnotation)
	}

	return config, nil
}

func (p *rewrite) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, rewriteAnnotation.Annotations)
}
