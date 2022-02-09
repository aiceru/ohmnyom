package errors

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func New(format string, a ...interface{}) error {
	return fmt.Errorf(format, a...)
}

func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

func GrpcError(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case *InvalidParamError:
		return status.Error(codes.Internal, err.Error())
	case *InvalidFormatError:
		return status.Error(codes.Internal, err.Error())
	case *NotFoundError:
		return status.Error(codes.NotFound, err.Error())
	case *AuthenticationError:
		return status.Error(codes.Unauthenticated, err.Error())
	case *UnimplementedError:
		return status.Error(codes.Unimplemented, err.Error())
	case *AlreadyExistsError:
		return status.Error(codes.AlreadyExists, err.Error())
	case *InternalError:
		return status.Error(codes.Internal, err.Error())
	}
	return status.Error(codes.Unknown, err.Error())
}

type InvalidParamError struct{ Err error }
type NotFoundError struct{ Err error }
type InvalidFormatError struct{ Err error }
type AuthenticationError struct{ Err error }
type UnimplementedError struct{ Err error }
type AlreadyExistsError struct{ Err error }
type InternalError struct{ Err error }
type NotSupportedError struct{ Err error }

func (e *InvalidParamError) Unwrap() error { return e.Err }
func (e *InvalidParamError) Error() string { return e.Err.Error() }
func NewInvalidParamError(format string, a ...interface{}) *InvalidParamError {
	return &InvalidParamError{Err: fmt.Errorf("invalid param: "+format, a...)}
}

func (e *NotFoundError) Unwrap() error { return e.Err }
func (e *NotFoundError) Error() string { return e.Err.Error() }
func NewNotFoundError(format string, a ...interface{}) *NotFoundError {
	return &NotFoundError{Err: fmt.Errorf("not found: "+format, a...)}
}

func (e *InvalidFormatError) Unwrap() error { return e.Err }
func (e *InvalidFormatError) Error() string { return e.Err.Error() }
func NewInvalidFormatError(format string, a ...interface{}) *InvalidFormatError {
	return &InvalidFormatError{Err: fmt.Errorf("invalid format: "+format, a...)}
}

func (e *AuthenticationError) Unwrap() error { return e.Err }
func (e *AuthenticationError) Error() string { return e.Err.Error() }
func NewAuthenticationError(format string, a ...interface{}) *AuthenticationError {
	return &AuthenticationError{Err: fmt.Errorf("auth error: "+format, a...)}
}

func (e *UnimplementedError) Unwrap() error { return e.Err }
func (e *UnimplementedError) Error() string { return e.Err.Error() }
func NewUnimplementedError(format string, a ...interface{}) *UnimplementedError {
	return &UnimplementedError{Err: fmt.Errorf("unimplemented: "+format, a...)}
}

func (e *AlreadyExistsError) Unwrap() error { return e.Err }
func (e *AlreadyExistsError) Error() string { return e.Err.Error() }
func NewAlreadyExistsError(format string, a ...interface{}) *AlreadyExistsError {
	return &AlreadyExistsError{Err: fmt.Errorf("already exists: "+format, a...)}
}

func (e *InternalError) Unwrap() error { return e.Err }
func (e *InternalError) Error() string { return e.Err.Error() }
func NewInternalError(format string, a ...interface{}) *InternalError {
	return &InternalError{Err: fmt.Errorf("internal error: "+format, a...)}
}
