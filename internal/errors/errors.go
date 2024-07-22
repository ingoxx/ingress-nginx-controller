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

func NewInvalidAnnotationContent(name string, val interface{}) error {
	return InvalidContentError{
		Name: fmt.Sprintf("the annotation %v does not contain a valid value (%v)", name, val),
	}
}
