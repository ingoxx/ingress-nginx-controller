package parser

import (
	"errors"
	"fmt"
	ingressv1 "github.com/Lxb921006/ingress-nginx-kubebuilder/api/v1"
	kerr "github.com/Lxb921006/ingress-nginx-kubebuilder/internal/errors"
	"regexp"
)

type AnnotationValidator func(string) error

func IsRegexPatternRegex(str string) bool {
	pattern := `^\/(?:.*\(.+\).*|.*\[[^\[\]]+\].*)`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func GetDnsRegex(str string) string {
	p := `([a0-z9]+\.)+([a-z]+)`
	matched := regexp.MustCompile(p)
	dns := matched.FindStringSubmatch(str)
	if len(dns) == 0 {
		return ""
	}

	return dns[0]
}

func IsTargetPathRegex(str string) bool {
	pattern := `^\/(\w+)|^/\$([0-9])$|^\$([0-9])$`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func IsNonEmptyParenthesesRegex(str string) bool {
	pattern := `^\/.*\(.+\).*`
	matched, _ := regexp.MatchString(pattern, str)
	return matched
}

func PassIsIp(target string) bool {
	pattern := `^(\d+).((\d+).){2}(\d+)$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(target)
}

func IsValidHost(host string) bool {
	pattern := `([a0-z9]+\.)+([a-z]+)`
	matched := regexp.MustCompile(pattern)
	if !matched.MatchString(host) {
		return false
	}
	return true
}

func IsAnnotationsPrefix(annotation string) bool {
	pattern := `^` + AnnotationsPrefix + "/"
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

func IsWeightPrefix(annotation string) bool {
	pattern := `^weight=[1-9](\d+)$`
	re := regexp.MustCompile(pattern)
	return re.FindStringIndex(annotation) != nil
}

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

func CheckAnnotationsKey(name string, ing *ingressv1.Ingress, field AnnotationFields) (string, error) {
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
