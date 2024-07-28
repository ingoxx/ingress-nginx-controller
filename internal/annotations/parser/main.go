package parser

import (
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultAnnotationsPrefix = "ingress.nginx.kubebuilder.io"
)

var (
	AnnotationsPrefix = DefaultAnnotationsPrefix
)

type AnnotationFields map[string]AnnotationConfig

type Annotation struct {
	Group       string
	Annotations AnnotationFields
}

type AnnotationConfig struct {
	Doc string
}

type IngressAnnotation interface {
	Parse(*ingressv1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}

func IsAnnotationsPrefix(annotation string) bool {
	pattern := `^` + AnnotationsPrefix + "/"
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

func TrimAnnotationPrefix(annotation string) string {
	return strings.TrimPrefix(annotation, AnnotationsPrefix+"/")
}

func GetAnnotationWithPrefix(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationsPrefix, suffix)
}

type ingAnnotations map[string]string

func (a ingAnnotations) parseString(name string) (string, error) {
	val, ok := a[name]
	if ok {
		if val == "" {
			return "", kerr.NewInvalidAnnotationContent(name, val)
		}

		return val, nil
	}

	return "", kerr.ErrMissingAnnotations
}

func (a ingAnnotations) parseStringSlice(name string) ([]string, error) {
	var data = make([]string, 0)
	val, ok := a[name]
	if ok {
		if val == "" {
			return data, kerr.NewInvalidAnnotationContent(name, val)
		}

		return data, nil
	}

	return data, kerr.ErrMissingAnnotations
}

func (a ingAnnotations) parseBool(name string) (bool, error) {
	val, ok := a[name]
	if ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			return false, kerr.NewInvalidAnnotationContent(name, val)
		}
		return b, nil
	}
	return false, kerr.ErrMissingAnnotations
}

func GetStringAnnotation(name string, ing *ingressv1.Ingress) (string, error) {
	key, err := CheckAnnotationsKey(name, ing)
	if err != nil {
		return "", err
	}
	return ingAnnotations(ing.GetAnnotations()).parseString(key)
}

func GetBoolAnnotations(name string, ing *ingressv1.Ingress) (bool, error) {
	key, err := CheckAnnotationsKey(name, ing)
	if err != nil {
		return false, err
	}
	return ingAnnotations(ing.GetAnnotations()).parseBool(key)
}
