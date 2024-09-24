package grammar

import (
	"errors"
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
)

var (
	ErrUnhandledInput = errors.New("unhandled input")
	ErrInvalidNumber  = errors.New("invalid number")
)

type Token uint

const (
	Number Token = iota + 1
	Plus
	Minus
	Mul
	Div
	Bra
	Ket
)

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

func BuildState(b state.Builder[Token]) []state.Update[Token] {
	return state.AsSlice[state.Update[Token]](
		// Spaces and tabs are omitted.
		b.Named("OmitSpaces").CheckRune(unicode.IsSpace).Repeat(state.CountBetween(1, math.MaxUint)).Omit(),
		// Parens
		b.Named("Ket").Rune(')').Emit(Ket).Break(),
		b.Named("Bra").Rune('(').Emit(Bra).State(b, BuildState),
		// Operators
		b.Named("Plus").Rune('+').Emit(Plus),
		b.Named("Minus").Rune('-').Emit(Minus),
		b.Named("Mul").Rune('*').Emit(Mul),
		b.Named("Div").Rune('/').Emit(Div),
		// Operands
		b.Named("Number").CheckRune(unicode.IsDigit).State(b, numberSubState).Optional().Emit(Number),
		// Error
		b.Named("UnhandledInput").Rest().Error(ErrUnhandledInput),
	)
}
