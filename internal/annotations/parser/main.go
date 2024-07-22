package parser

import (
	serr "errors"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	DefaultAnnotationsPrefix = "nginx.ingress.kubebuilder.io"
)

var (
	AnnotationsPrefix = DefaultAnnotationsPrefix
)

type AnnotationFields map[string]string

type Annotation struct {
	Group       string
	Annotations AnnotationFields
}

type IngressAnnotation interface {
	Parse(*ingressv1.Ingress) (interface{}, error)
	Validate(map[string]string) error
}

func TrimAnnotationPrefix(annotation string) string {
	return strings.TrimPrefix(annotation, AnnotationsPrefix+"/")
}

func GetAnnotationWithPrefix(suffix string) string {
	return fmt.Sprintf("%v/%v", AnnotationsPrefix, suffix)
}

func GetBackendHost(host string) (string, error) {
	pattern := `((\w+).)+`
	re := regexp.MustCompile(pattern)
	domain := re.FindStringSubmatch(host)
	if len(domain) == 0 {
		return "", serr.New("host filed cannot be empty")
	}

	return domain[0], nil
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
