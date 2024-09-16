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
	return lex.AnyString("\n", "\r\n").
		Emit(lexer.NL).Done()
}

func skipSpaces(lex *lexer.Context[testMessageType]) bool {
	return lex.While(unicode.IsSpace).Skip().Done()
}
func number(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(unicode.IsDigit).
		OptionallyWhile(unicode.IsDigit).
		Emit2(Number).
		Done()
}
func negativeNumber(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(lexer.Rune('-')).
		Fn(unicode.IsDigit).
		OptionallyWhile(unicode.IsDigit).
		Emit2(Number).
		Done()
}
func minus(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(lexer.Rune('-')).
		Emit2(Minus).
		Done()
}
func identifier(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(unicode.IsLetter).
		OptionallyWhile(func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }).
		Emit2(Identifier).
		Done()
}
func stringData(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(lexer.Rune('"')).
		OptionallyWhile(lexer.Escape(lexer.Rune('\\'), func(r rune) bool { return r != '"' }).Accept).
		Fn(lexer.Rune('"')).
		Emit2(String).
		Done()
}
func scopeContext(lex *lexer.Context[testMessageType]) bool {
	return lex.Fn(lexer.Rune('(')).If(func(lex *lexer.Context[testMessageType]) {
		switch {
		case skipSpaces(lex):
		case lex.Fn(lexer.Rune(')')).Emit(lexer.User, Ket).Done():
			lex.Break()
		case lex.Fn(lexer.Rune(',')).Emit(lexer.User, Comma).Done():
		case negativeNumber(lex):
		case minus(lex):
		case number(lex):
		case identifier(lex):
		case stringData(lex):
		case scopeContext(lex):
		default:
			lex.SetError(ErrUnexpectedState)
		}
	}).Emit2(Bra).Done()
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
