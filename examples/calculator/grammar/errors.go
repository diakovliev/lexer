package grammar

import "errors"

var (
	// ErrInvalidExpression is returned when the lexer encounters an invalid expression.
	ErrInvalidExpression = errors.New("invalid expression")
	// ErrInvalidNumber is returned when the lexer encounters an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrUnexpectedBra is returned when the lexer encounters an unexpected '(' character. It means
	// that expression is reached max scopes depth.
	ErrUnexpectedBra = errors.New("unexpected '('")
	// ErrUnexpectedKet is returned when the lexer encounters an unexpected ')' character.
	ErrUnexpectedKet = errors.New("unexpected ')'")
	// ErrDisabledHistory is returned when the lexer tries to use history and it's disabled.
	ErrDisabledHistory = errors.New("history disabled")
	// ErrInvalidIdentifier is returned when the lexer encounters an invalid identifier.
	ErrInvalidIdentifier = errors.New("invalid identifier")
)
