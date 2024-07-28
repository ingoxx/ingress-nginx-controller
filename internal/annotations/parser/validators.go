package parser

import (
	"errors"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"regexp"
)

func CheckAnnotations(annotations map[string]string, config AnnotationFields) error {
	var err error
	for annotation := range annotations {
		if !IsAnnotationsPrefix(annotation) {
			continue
		}

		annPure := TrimAnnotationPrefix(annotation)
		if cfg, ok := config[annPure]; ok && cfg.Doc == "" {
			err = errors.Join(err, fmt.Errorf("annotation %s have no description", annotation))
		}
	}

	return err
}

func CheckAnnotationsKey(name string, ing *ingressv1.Ingress) (string, error) {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return "", kerr.ErrMissingAnnotations
	}

	annotationFullName := GetAnnotationWithPrefix(name)
	annotationValue := ing.GetAnnotations()[annotationFullName]

	if annotationValue == "" {
		return "", kerr.NewValidationError(annotationFullName)
	}

	return annotationFullName, nil
}

func PassIsIp(target string) bool {
	pattern := `^(\d+).((\d+).){2}(\d+)$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(target)
}
