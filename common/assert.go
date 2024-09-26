package common

import (
	"errors"
	"fmt"
)

const (
	assertionError = "ASSERTION FAILED"
)

// AssertionError represents an error that occurs when an assertion fails. It contains a message and an optional underlying cause.
type AssertionError struct {
	Cause   error
	Message string
}

// newAssertionError creates a new AssertionError with the given message and optional underlying cause. The returned error will be of type *AssertionError.
func newAssertionError(message string, err error) *AssertionError {
	return &AssertionError{
		Cause:   err,
		Message: message,
	}
}

// Error implements the error interface. It returns a string describing the assertion error. If there is an underlying cause, it will be included in the returned message.
func (e *AssertionError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: message: '%s'", assertionError, e.Message)
	}
	return fmt.Sprintf("%s: cause: '%v' message: '%s'", assertionError, e.Cause, e.Message)
}

// Unwrap implements the errors.Unwrap interface. It allows to access the underlying cause of the error.
func (e *AssertionError) Unwrap() error {
	return e.Cause
}

// AssertNil asserts that the given value is nil. If it isn't, an assertion error will be raised with the given message.
func AssertNil(value any, message string) {
	if value == nil {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertNotNil asserts that the given value is not nil. If it is, an assertion error will be raised with the given message.
func AssertNotNil(value any, message string) {
	if value != nil {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertNilPtr asserts that the given pointer is nil. If it isn't, an assertion error will be raised with the given message.
func AssertNilPtr[T any](value *T, message string) {
	if value == nil {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertNotNilPtr asserts that the given pointer is not nil. If it is, an assertion error will be raised with the given message.
func AssertNotNilPtr[T any](value *T, message string) {
	if value != nil {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertTrue asserts that the given condition is true. If it isn't, an assertion error will be raised with the given message.
func AssertTrue(cond bool, message string) {
	if cond {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertFalse asserts that the given condition is false. If it isn't, an assertion error will be raised with the given message.
func AssertFalse(cond bool, message string) {
	if !cond {
		return
	}
	panic(newAssertionError(message, nil))
}

// AssertTrueFn asserts that the given function returns true. If it doesn't, an assertion error will be raised with the given message.
func AssertTrueFn(cond func() bool, message string) {
	AssertTrue(cond(), message)
}

// AssertFalseFn asserts that the given function returns false. If it doesn't, an assertion error will be raised with the given message.
func AssertFalseFn(cond func() bool, message string) {
	AssertFalse(cond(), message)
}

// AssertError asserts that the given error is not nil. If it's nil, an assertion error will be raised with the given message and cause.
func AssertError(err error, message string) {
	if err != nil {
		return
	}
	panic(newAssertionError(message, err))
}

// AssertNoError asserts that the given error is nil. If it's not, an assertion error will be raised with the given message and cause.
func AssertNoError(err error, message string) {
	if err == nil {
		return
	}
	panic(newAssertionError(message, err))
}

// AssertErrorIs asserts that the given error is of the given type. If it's not, an assertion error will be raised with the given message and cause.
func AssertErrorIs(err error, expected error, message string) {
	if errors.Is(err, expected) {
		return
	}
	panic(newAssertionError(message, fmt.Errorf("expected %v to be of type %T", err, expected)))
}

// AssertErrorNotIs asserts that the given error is not of the given type. If it's, an assertion error will be raised with the given message and cause.
func AssertErrorIsAnyFrom(err error, expected []error, message string) {
	for _, e := range expected {
		if errors.Is(err, e) {
			return
		}
	}
	panic(newAssertionError(message, fmt.Errorf("expected %v to be of any type in %T", err, expected)))
}

// AssertErrorNotIs asserts that the given error is not of the given type. If it's, an assertion error will be raised with the given message and cause.
func AssertErrorNotIs(err error, unexpected error, message string) {
	if !errors.Is(err, unexpected) {
		return
	}
	panic(newAssertionError(message, fmt.Errorf("expected %v to not be of type %T", err, unexpected)))
}

// AssertNoErrorOrIs asserts that the given error is nil or of the given type. If it's not, an assertion error will be raised with the given message and cause.
func AssertNoErrorOrIs(err error, expected error, message string) {
	if err == nil || errors.Is(err, expected) {
		return
	}
	panic(newAssertionError(message, fmt.Errorf("expected %v to be of type %T or nil", err, expected)))
}
