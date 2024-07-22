package parser

import (
	"errors"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
)

func CheckAnnotations(annotations map[string]string, config AnnotationFields) error {
	var err error
	for annotation := range annotations {
		annPure := TrimAnnotationPrefix(annotation)
		if _, ok := config[annPure]; !ok {
			err = errors.Join(err, fmt.Errorf("annotation %s is too risky for environment", annotation))
		}
	}

	return err
}

func CheckAnnotationsKey(name string, ing *ingressv1.Ingress) (string, error) {
	if ing == nil || len(ing.GetAnnotations()) == 0 {
		return "", kerr.ErrMissingAnnotations
	}

	annotationFullName := GetAnnotationWithPrefix(name)
	if annotationFullName == "" {
		return "", kerr.ErrInvalidAnnotationName
	}

	annotationValue := ing.GetAnnotations()[annotationFullName]
	if annotationValue == "" {
		return "", kerr.NewValidationError(annotationFullName)
	}

	return annotationFullName, nil
}
