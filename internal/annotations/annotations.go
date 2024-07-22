package annotations

import (
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/ipallowlist"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/ipdenylist"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/parser"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/proxy"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/redirect"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/resolver"
	"github.com/Lxb921006/ingress-nginx-kubebuilder/internal/annotations/rewrite"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"github.com/imdario/mergo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type Extractor struct {
	annotations map[string]parser.IngressAnnotation
}

type IngressAnnotations struct {
	Ingress           *ingressv1.Ingress `json:"ingress"`
	ParsedAnnotations *Ingress           `json:"parsed_annotations"`
}

type Ingress struct {
	metav1.ObjectMeta
	Proxy     proxy.Config
	Rewrite   rewrite.Config
	Redirect  redirect.Config
	AllowList ipallowlist.SourceRange
	DenyList  ipdenylist.SourceRange
}

func NewAnnotationExtractor(r resolver.Resolver) *Extractor {
	return &Extractor{
		map[string]parser.IngressAnnotation{
			"Proxy":     proxy.NewParser(r),
			"Redirect":  redirect.NewParser(r),
			"AllowList": ipallowlist.NewParser(r),
			"DenyList":  ipdenylist.NewParser(r),
			"Rewrite":   rewrite.NewParser(r),
		},
	}
}

func (e Extractor) Extract(ing *ingressv1.Ingress) (*Ingress, error) {
	pia := &Ingress{
		ObjectMeta: ing.ObjectMeta,
	}

	data := make(map[string]interface{})
	for name, annotationParser := range e.annotations {
		if err := annotationParser.Validate(ing.GetAnnotations()); err != nil {
			return nil, kerr.NewRiskyAnnotations(name)
		}
		val, err := annotationParser.Parse(ing)
		klog.V(5).InfoS("Parsing Ingress annotation", "name", name, "ingress", klog.KObj(ing), "value", val)
		if err != nil {
			if kerr.IsValidationError(err) {
				klog.ErrorS(err, "ingress contains invalid annotation value")
				return nil, err
			}

			if kerr.IsMissingAnnotations(err) {
				continue
			}
		}
		if val != nil {
			data[name] = val
		}
	}

	err := mergo.MapWithOverwrite(pia, data)
	if err != nil {
		klog.ErrorS(err, "unexpected error merging extracted annotations")
	}

	return pia, nil
}
