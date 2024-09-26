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
	// Cause is the optional underlying cause of this AssertionError. This may be another error or a string. If not nil, it will be included in the returned message.
	Cause error
	// Message is the message of this AssertionError. This should describe what failed and why.
	Message string
}

// newAssertionError creates a new AssertionError with the given message and optional underlying cause. The returned error will be of type *AssertionError.
func newAssertionError(cause error, message string, args ...any) *AssertionError {
	return &AssertionError{
		Cause:   cause,
		Message: fmt.Sprintf(message, args...),
	}
}

// Error implements the error interface. It returns a string describing the assertion error. If there is an underlying cause, it will be included in the returned message.
func (e *AssertionError) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", assertionError, e.Message)
	}
	return fmt.Sprintf("%s[%v]: %s", assertionError, e.Cause, e.Message)
}

// Unwrap implements the errors.Unwrap interface. It allows to access the underlying cause of the error.
func (e *AssertionError) Unwrap() error {
	return e.Cause
}

// AssertUnreachable asserts that this code path should never be reached. It panics with a new AssertionError containing the given message.
// This function should only be used in unreachable code paths. If it is called, an assertion error will be raised.
func AssertUnreachable(message string, args ...any) {
	panic(newAssertionError(nil, message, args...))
}

// AssertNil asserts that the given value is nil. If it isn't, an assertion error will be raised with the given message.
func AssertNil(value any, message string, args ...any) {
	if value == nil {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertNotNil asserts that the given value is not nil. If it is, an assertion error will be raised with the given message.
func AssertNotNil(value any, message string, args ...any) {
	if value != nil {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertNilPtr asserts that the given pointer is nil. If it isn't, an assertion error will be raised with the given message.
func AssertNilPtr[T any](value *T, message string, args ...any) {
	if value == nil {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertNotNilPtr asserts that the given pointer is not nil. If it is, an assertion error will be raised with the given message.
func AssertNotNilPtr[T any](value *T, message string, args ...any) {
	if value != nil {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertTrue asserts that the given condition is true. If it isn't, an assertion error will be raised with the given message.
func AssertTrue(cond bool, message string, args ...any) {
	if cond {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertFalse asserts that the given condition is false. If it isn't, an assertion error will be raised with the given message.
func AssertFalse(cond bool, message string, args ...any) {
	if !cond {
		return
	}
	panic(newAssertionError(nil, message, args...))
}

// AssertTrueFn asserts that the given function returns true. If it doesn't, an assertion error will be raised with the given message.
func AssertTrueFn(cond func() bool, message string, args ...any) {
	AssertTrue(cond(), message, args...)
}

// AssertFalseFn asserts that the given function returns false. If it doesn't, an assertion error will be raised with the given message.
func AssertFalseFn(cond func() bool, message string, args ...any) {
	AssertFalse(cond(), message, args...)
}

// AssertError asserts that the given error is not nil. If it's nil, an assertion error will be raised with the given message and cause.
func AssertError(err error, message string, args ...any) {
	if err != nil {
		return
	}
	panic(newAssertionError(err, message, args...))
}

// AssertNoError asserts that the given error is nil. If it's not, an assertion error will be raised with the given message and cause.
func AssertNoError(err error, message string, args ...any) {
	if err == nil {
		return
	}
	panic(newAssertionError(err, message, args...))
}

// AssertErrorIs asserts that the given error is of the given type. If it's not, an assertion error will be raised with the given message and cause.
func AssertErrorIs(err error, expected error, message string, args ...any) {
	if errors.Is(err, expected) {
		return
	}
	panic(newAssertionError(fmt.Errorf("expected %v to be of type %T", err, expected), message, args...))
}

// AssertErrorNotIs asserts that the given error is not of the given type. If it's, an assertion error will be raised with the given message and cause.
func AssertErrorIsAnyFrom(err error, expected []error, message string, args ...any) {
	for _, e := range expected {
		if errors.Is(err, e) {
			return
		}
	}
	panic(newAssertionError(fmt.Errorf("expected %v to be of any type in %T", err, expected), message, args...))
}

// AssertErrorNotIs asserts that the given error is not of the given type. If it's, an assertion error will be raised with the given message and cause.
func AssertErrorNotIs(err error, unexpected error, message string, args ...any) {
	if !errors.Is(err, unexpected) {
		return
	}
	panic(newAssertionError(fmt.Errorf("expected %v to not be of type %T", err, unexpected), message, args...))
}

// AssertNoErrorOrIs asserts that the given error is nil or of the given type. If it's not, an assertion error will be raised with the given message and cause.
func AssertNoErrorOrIs(err error, expected error, message string, args ...any) {
	if err == nil || errors.Is(err, expected) {
		return
	}
	panic(newAssertionError(fmt.Errorf("expected %v to be of type %T or nil", err, expected), message, args...))
}
