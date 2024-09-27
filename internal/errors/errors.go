package errors

import (
	"errors"
	"fmt"
)

var (
	ErrMissingAnnotations    = errors.New("ingress rule without annotations")
	ErrInvalidAnnotationName = errors.New("invalid annotation name")
)

type RiskyAnnotationError struct {
	Reason error
}

func (e RiskyAnnotationError) Error() string {
	return e.Reason.Error()
}

func IsMissingAnnotations(e error) bool {
	return errors.Is(e, ErrMissingAnnotations)
}

func NewRiskyAnnotations(name string) error {
	return RiskyAnnotationError{
		Reason: fmt.Errorf("annotation group %s contains risky annotation based on ingress configuration", name),
	}
}

type ValidationError struct {
	Reason error
}

func (e ValidationError) Error() string {
	return e.Reason.Error()
}

func NewValidationError(annotation string) error {
	return ValidationError{
		Reason: fmt.Errorf("annotation %s contains invalid value", annotation),
	}
}

func IsValidationError(e error) bool {
	var validationError ValidationError
	ok := errors.As(e, &validationError)
	return ok
}

type InvalidContentError struct {
	Name string
}

func (e InvalidContentError) Error() string {
	return e.Name
}

func NewInvalidContent(name string, val interface{}) error {
	return InvalidContentError{
		Name: fmt.Sprintf("the annotation %v does not contain a valid value (%v)", name, val),
	}
}

type InvalidIngressContentError struct {
	Name string
}

func (e InvalidIngressContentError) Error() string {
	return e.Name
}

func NewInvalidIngressContent(name string, val interface{}) error {
	return InvalidIngressContentError{
		Name: fmt.Sprintf("the ingress %v does not contain a valid value (%v)", name, val),
	}
}

func IsInvalidIngressContentError(e error) bool {
	var invalidIngressContentError InvalidIngressContentError
	ok := errors.As(e, &invalidIngressContentError)
	return ok
}

type InvalidAnnotationContentError struct {
	Name string
}

func (e InvalidAnnotationContentError) Error() string {
	return e.Name
}

func IsInvalidAnnotationsContentError(e error) bool {
	var invalidAnnotationsContentError InvalidAnnotationContentError
	ok := errors.As(e, &invalidAnnotationsContentError)
	return ok
}

func NewInvalidAnnotationsContentError(name string, val interface{}) error {
	return InvalidAnnotationContentError{
		Name: fmt.Sprintf("the annotation %v does not contain a valid value (%v)", name, val),
	}
}

type MissResourcesError struct {
	Name string
}

func (e MissResourcesError) Error() string {
	return e.Name
}

func IsMissResourcesError(e error) bool {
	var resourceNotFoundError MissResourcesError
	ok := errors.As(e, &resourceNotFoundError)
	return ok
}

func NewIsMissResourcesError(name string) error {
	return MissResourcesError{
		Name: fmt.Sprintf("service resouce %s not found.", name),
	}
}
