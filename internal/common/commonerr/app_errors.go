package commonerr

import (
	"errors"
	"fmt"
)

type ErrorType struct {
	t string
}

var (
	ErrorTypeUnknown        = ErrorType{"unknown"}
	ErrorTypeIncorrectInput = ErrorType{"incorrect-input"}
	ErrorTypeNotFound       = ErrorType{"not-found"}
)

type AppError struct {
	messagef  string
	args      []interface{}
	errorType ErrorType
}

func (e AppError) Error() string {
	return fmt.Sprintf(e.messagef, e.args...)
}

func (e AppError) ErrorType() ErrorType {
	return e.errorType
}

func (e AppError) Is(target error) bool {
	var appErr AppError

	if !errors.As(target, &appErr) {
		return false
	}

	return appErr.ErrorType() == e.ErrorType()
}

func IsAppError(err error) bool {
	var appErr AppError

	return errors.As(err, &appErr)
}

func NewUnknownAppErrorf(messagef string, args ...interface{}) AppError {
	return AppError{
		messagef:  messagef,
		args:      args,
		errorType: ErrorTypeUnknown,
	}
}

func NewIncorrectInputError(messagef string, args ...interface{}) AppError {
	return AppError{
		messagef:  messagef,
		args:      args,
		errorType: ErrorTypeIncorrectInput,
	}
}

func NewNotFoundError(messagef string, args ...interface{}) AppError {
	return AppError{
		messagef:  messagef,
		args:      args,
		errorType: ErrorTypeNotFound,
	}
}
