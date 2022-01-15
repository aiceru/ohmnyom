package errors

import (
	"errors"
	"fmt"
)

func New(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

type InvalidParamError struct {
	Err error
	// trace *sentry.Stacktrace
}
type NotFoundError struct {
	Err error
}
type InvalidFormatError struct {
	Err error
}

func (e *InvalidParamError) Unwrap() error {
	return e.Err
}

func (e *InvalidParamError) Error() string {
	return e.Err.Error()
}

// func (e *InvalidParamError) GetTrace() *sentry.Stacktrace {
// 	return e.trace
// }

func NewInvalidParamError(format string, a ...interface{}) *InvalidParamError {
	return &InvalidParamError{
		Err: fmt.Errorf("invalid param: "+format, a...),
		// trace: sentry.NewStacktrace(2),
	}
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func (e *NotFoundError) Error() string {
	return e.Err.Error()
}

func NewNotFoundError(format string, a ...interface{}) *NotFoundError {
	return &NotFoundError{
		Err: fmt.Errorf("not found : "+format, a...),
	}
}

func (e *InvalidFormatError) Unwrap() error {
	return e.Err
}

func (e *InvalidFormatError) Error() string {
	return e.Err.Error()
}

func NewInvalidFormatError(format string, a ...interface{}) *InvalidFormatError {
	return &InvalidFormatError{
		Err: fmt.Errorf("invalid format : "+format, a...),
	}
}
