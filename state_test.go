package lexer_test

import (
	"errors"
	"unicode"

	"github.com/diakovliev/lexer"
)

type (
	testMessageType int
	testState       struct{}
)

const (
	Identifier testMessageType = iota
	Plus
	Minus
	Number
	String
	Bra
	Ket
	Comma
	Last
)

var (
	ErrUnexpectedState = errors.New("unexpected state")
)

func multiline(lex *lexer.Context[testMessageType]) bool {
	return lex.AcceptAnyStringFrom("\n", "\r\n").
		Emit(lexer.NL).Done()
}

func skipSpaces(lex *lexer.Context[testMessageType]) bool {
	return lex.AcceptWhile(unicode.IsSpace).Skip().Done()
}
func number(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(unicode.IsDigit).
		OptionallyAcceptWhile(unicode.IsDigit).
		Emit(lexer.User, Number).
		Done()
}
func negativeNumber(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(lexer.Rune('-')).
		Accept(unicode.IsDigit).
		OptionallyAcceptWhile(unicode.IsDigit).
		Emit(lexer.User, Number).
		Done()
}
func minus(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(lexer.Rune('-')).
		Emit(lexer.User, Minus).
		Done()
}
func identifier(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(unicode.IsLetter).
		OptionallyAcceptWhile(func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }).
		Emit(lexer.User, Identifier).
		Done()
}
func stringData(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(lexer.Rune('"')).
		OptionallyAcceptWhile(lexer.Escape(lexer.Rune('\\'), func(r rune) bool { return r != '"' }).Accept).
		Accept(lexer.Rune('"')).
		Emit(lexer.User, String).
		Done()
}
func scopeContext(lex *lexer.Context[testMessageType]) bool {
	return lex.Accept(lexer.Rune('(')).AcceptContext(func(lex *lexer.Context[testMessageType]) {
		switch {
		case skipSpaces(lex):
		case lex.Accept(lexer.Rune(')')).Emit(lexer.User, Ket).Done():
			lex.Break()
		case lex.Accept(lexer.Rune(',')).Emit(lexer.User, Comma).Done():
		case negativeNumber(lex):
		case minus(lex):
		case number(lex):
		case identifier(lex):
		case stringData(lex):
		case scopeContext(lex):
		default:
			lex.SetError(ErrUnexpectedState)
		}
	}).Emit(lexer.User, Bra).Done()
}
func testInitialState(lex *lexer.Context[testMessageType]) {
	switch {
	case skipSpaces(lex):
	case negativeNumber(lex):
	case minus(lex):
	case number(lex):
	case identifier(lex):
	case stringData(lex):
	case scopeContext(lex):
	default:
		lex.SetError(ErrUnexpectedState)
	}
}
