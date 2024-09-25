package grammar

import (
	"context"
	"errors"
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
	"github.com/diakovliev/lexer/xio"
)

var (
	// ErrUnhandledInput is returned when the lexer encounters an unhandled input.
	ErrUnhandledInput = errors.New("unhandled input")
	// ErrInvalidNumber is returned when the lexer encounters an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrUnexpectedBra is returned when the lexer encounters an unexpected '(' character. It means
	// that expression is reached max scopes depth.
	ErrUnexpectedBra = errors.New("unexpected '('")
	// ErrUnexpectedKet is returned when the lexer encounters an unexpected ')' character.
	ErrUnexpectedKet = errors.New("unexpected ')'")
)

var (
	plusMinus = state.Or(
		state.IsRune('+'),
		state.IsRune('-'),
	)

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

func ketState(root bool) (ket func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		base := b.Named("Ket").Rune(')')
		if root {
			return base.Error(ErrUnexpectedKet)
		}
		return base.Emit(Ket).Break()
	}
}

func braState(depth uint) func(b state.Builder[Token]) *state.Chain[Token] {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		base := b.Named("Bra").Rune('(')
		if depth > 0 {
			return base.Emit(Bra).State(b, newState(false, depth-1))
		}
		return base.Error(ErrUnexpectedBra)
	}
}

func signedNumberGuard(ctx context.Context, _ xio.State) (err error) {
	history := state.GetHistory[Token](ctx).Get()
	if len(history) == 0 {
		return
	}
	if history[len(history)-1].Token == Number {
		err = state.ErrRollback
	}
	return
}

func numberSubState(b state.Builder[Token]) []state.Update[Token] {
	// consume all digits
	digits := func(b state.Builder[Token]) *state.Chain[Token] {
		return b.CheckRune(unicode.IsDigit).Repeat(state.CountBetween(0, math.MaxUint))
	}
	return state.AsSlice[state.Update[Token]](
		digits(b).FollowedByCheckRune(allTerms).Break(),
		// if followed by non known term, emit error
		digits(b).FollowedByCheckNotRune(unicode.IsDigit).Rest().Error(ErrInvalidNumber),
		// otherwise, break
		digits(b).Break(),
	)
}

func numberState(name string, signed bool) (number func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) (state *state.Chain[Token]) {
		state = b.Named(name)
		if signed {
			state = state.CheckRune(plusMinus).Optional().Tap(signedNumberGuard)
		}
		state = state.CheckRune(unicode.IsDigit).State(b, numberSubState).Optional().Emit(Number)
		return
	}
}

// newState returns a new state machine for parsing tokens from the input string.
func newState(root bool, maxScopesDepth uint) func(b state.Builder[Token]) []state.Update[Token] {
	return func(b state.Builder[Token]) []state.Update[Token] {
		return state.AsSlice[state.Update[Token]](
			// Spaces and tabs are omitted.
			b.Named("OmitSpaces").CheckRune(unicode.IsSpace).Repeat(state.CountBetween(1, math.MaxUint)).Omit(),
			// Parens with max depth
			ketState(root)(b),
			braState(maxScopesDepth-1)(b),
			// Operands
			numberState("SignedNumber", true)(b),
			numberState("UnsignedNumber", false)(b),
			// Operators
			b.Named("Plus").Rune('+').Emit(Plus),
			b.Named("Minus").Rune('-').Emit(Minus),
			b.Named("Mul").Rune('*').Emit(Mul),
			b.Named("Div").Rune('/').Emit(Div),
			// Error
			b.Named("UnhandledInput").Rest().Error(ErrUnhandledInput),
		)
	}
}
