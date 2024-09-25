package grammar

import (
	"errors"
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
)

var (
	// ErrUnhandledInput is returned when the lexer encounters an unhandled input.
	ErrUnhandledInput = errors.New("unhandled input")
	// ErrInvalidNumber is returned when the lexer encounters an invalid number.
	ErrInvalidNumber = errors.New("invalid number")
	// ErrUnexpectedKet is returned when the lexer encounters an unexpected ')' character.
	ErrUnexpectedKet = errors.New("unexpected ')'")
)

// Token is a lexer token.
type Token uint

const (
	// Number is a number token.
	Number Token = iota
	// Plus is a plus token.
	Plus
	// Minus is a minus token.
	Minus
	// Mul is a multiplication token.
	Mul
	// Div is a division token.
	Div
	// Bra is an opening bracket token.
	Bra
	// Ket is a closing bracket token.
	Ket
)

// String returns the string representation of a token.
func (t Token) String() string {
	switch t {
	case Number:
		return "Number"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Mul:
		return "Mul"
	case Div:
		return "Div"
	case Bra:
		return "Bra"
	case Ket:
		return "Ket"
	default:
		panic("unreachable")
	}
}

var allTerms = state.Or(
	unicode.IsSpace,
	state.IsRune('+'),
	state.IsRune('-'),
	state.IsRune('*'),
	state.IsRune('/'),
	state.IsRune(')'),
	state.IsRune('('),
)

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

// New returns a new state machine for parsing tokens from the input string.
func New(root bool) func(b state.Builder[Token]) []state.Update[Token] {
	var ket func(b state.Builder[Token]) *state.Chain[Token]
	if root {
		ket = func(b state.Builder[Token]) *state.Chain[Token] {
			return b.Named("Ket").Rune(')').Error(ErrUnexpectedKet)
		}
	} else {
		ket = func(b state.Builder[Token]) *state.Chain[Token] {
			return b.Named("Ket").Rune(')').Emit(Ket).Break()
		}
	}
	return func(b state.Builder[Token]) []state.Update[Token] {
		return state.AsSlice[state.Update[Token]](
			// Spaces and tabs are omitted.
			b.Named("OmitSpaces").CheckRune(unicode.IsSpace).Repeat(state.CountBetween(1, math.MaxUint)).Omit(),
			// Parens
			ket(b),
			b.Named("Bra").Rune('(').Emit(Bra).State(b, New(false)),
			// Operands
			// b.Named("Number").Rune('-').Optional().CheckRune(unicode.IsDigit).State(b, numberSubState).Optional().Emit(Number),
			b.Named("Number").CheckRune(unicode.IsDigit).State(b, numberSubState).Optional().Emit(Number),
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
