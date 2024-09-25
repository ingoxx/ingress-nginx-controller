package weight

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
	useWeightAnnotation = "use-weight"
	setWeightAnnotation = "set-weight"
)

var weightAnnotation = parser.Annotation{
	Group: "allowCos",
	Annotations: parser.AnnotationFields{
		useWeightAnnotation: {
			Doc: "use weight, e.g: `true or false`, required",
		},
		setWeightAnnotation: {
			Doc: "target backend, e.g: `svc-name1:weight=80,svc-name2:weight=20...`, required",
		},
	},
}

type BackendWeight struct {
	SvcList  []string `json:"svc-list"`
	Upstream string   `json:"upstream"`
	*Config
}

type Config struct {
	UseWeight bool   `json:"weight"`
	SetWeight string `json:"set-weight"`
}

type weight struct {
	r resolver.Resolver
}

func NewParser(r resolver.Resolver) parser.IngressAnnotation {
	return &weight{
		r: r,
	}
}

func (r *weight) Parse(ing *ingressv1.Ingress) (interface{}, error) {
	var err error
	bw := &BackendWeight{
		Config: &Config{},
	}
	bw.UseWeight, err = parser.GetBoolAnnotations(useWeightAnnotation, ing, weightAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to false", useWeightAnnotation)
		}
	}

	bw.SetWeight, err = parser.GetStringAnnotation(setWeightAnnotation, ing, weightAnnotation.Annotations)
	if err != nil {
		if errors.IsValidationError(err) {
			klog.Warningf("%s is invalid, defaulting to empty", setWeightAnnotation)
		}
	}

	if err = r.check(ing, bw); err != nil {
		return nil, err
	}

	return bw, nil
}

func (r *weight) check(ing *ingressv1.Ingress, config *BackendWeight) error {
	if config.UseWeight {
		rules := ing.Spec.Rules
		for _, r := range rules {
			var path string
			for _, p := range r.HTTP.Paths {
				if path == "" {
					path = p.Path
				}

				if path != p.Path {
					return fmt.Errorf("due to the use of traffic allocation, the path field value should be the same")
				}
			}
		}

		svcList := strings.Split(config.SetWeight, ",")
		if len(svcList) < 1 {
			return fmt.Errorf("at least two services are required to use the traffic allocation function")
		}

		var upstreamName string

		for _, alias := range svcList {
			alias = strings.TrimSpace(alias)
			if alias == "" {
				continue
			}

			val := strings.Split(alias, ":")
			svc, err := r.r.GetService(val[0])
			if err != nil {
				return err
			}

			svcPort := r.r.GetSvcPort(val[0])
			if svcPort == 0 {
				return fmt.Errorf("svc %s not found", val[0])
			}

			if !parser.IsWeightPrefix(val[1]) {
				return errors.NewInvalidAnnotationContent(setWeightAnnotation, config.SetWeight)
			}

			if upstreamName == "" {
				upstreamName += svc.Name
			} else {
				upstreamName += "-" + svc.Name
			}

			config.SvcList = append(config.SvcList, fmt.Sprintf("%s.%s.svc:%d %s", svc.Name, svc.Namespace, svcPort, val[1]))

		}

		if config.Upstream == "" {
			config.Upstream = fmt.Sprintf("%s-%s-%s", upstreamName, ing.Name, ing.Namespace)
		}

	}

	return nil
}

func (r *weight) Validate(anns map[string]string) error {
	return parser.CheckAnnotations(anns, weightAnnotation.Annotations)
}
