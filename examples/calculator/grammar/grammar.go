package grammar

import (
	"errors"
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
)

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
)

var (
	allTerms = state.Or(
		unicode.IsSpace,
		state.IsRune('+'),
		state.IsRune('-'),
		state.IsRune('*'),
		state.IsRune('/'),
		state.IsRune(')'),
		state.IsRune('('),
	)
)

func ketState(name string, root bool) (ket func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		base := b.Named(name).Rune(')')
		if root {
			return base.Error(ErrUnexpectedKet)
		}
		return base.Emit(Ket).Break()
	}
}

func braState(name string, depth uint) func(b state.Builder[Token]) *state.Chain[Token] {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		base := b.Named(name).Rune('(')
		if depth > 0 {
			return base.Emit(Bra).State(b, newState(false, depth-1))
		}
		return base.Error(ErrUnexpectedBra)
	}
}

// newState returns a new state machine for parsing tokens from the input string.
func newState(root bool, maxScopesDepth uint) func(b state.Builder[Token]) []state.Update[Token] {
	return func(b state.Builder[Token]) []state.Update[Token] {
		base := []state.Update[Token]{
			// Spaces and tabs are omitted.
			b.Named("OmitSpaces").RuneCheck(unicode.IsSpace).Repeat(state.CountBetween(1, math.MaxUint)).Omit(),
			// Parens with max depth
			braState("Bra", maxScopesDepth-1)(b),
			ketState("Ket", root)(b),
		}
		numbers := []state.Update[Token]{}
		for _, numberStateBuilder := range numberStateBuilders {
			numbers = append(numbers, numberStateBuilder.build(b))
		}
		rest := []state.Update[Token]{
			// Operators
			b.Named("Plus").Rune('+').Emit(Plus),
			b.Named("Minus").Rune('-').Emit(Minus),
			b.Named("Mul").Rune('*').Emit(Mul),
			b.Named("Div").Rune('/').Emit(Div),
			// Error
			b.Named("InvalidExpression").Rest().Error(ErrInvalidExpression),
		}

		// combine all
		all := []state.Update[Token]{}
		all = append(all, base...)
		all = append(all, numbers...)
		all = append(all, rest...)
		return all
	}
}
