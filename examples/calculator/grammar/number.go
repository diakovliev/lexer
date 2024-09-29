package grammar

import (
	"context"
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
	"github.com/diakovliev/lexer/xio"
)

var (
	allValueTokens = map[Token]bool{
		BinNumber:   true,
		OctNumber:   true,
		DecNumber:   true,
		HexNumber:   true,
		BinFraction: true,
		OctFraction: true,
		DecFraction: true,
		HexFraction: true,
		Identifier:  true,
	}

	plusMinus = state.Or(
		state.IsRune('+'),
		state.IsRune('-'),
	)

	// Numbers misc
	IsNumberDot = state.IsRune('.')
	isHexDigit  = state.Or(
		unicode.IsDigit,
		state.IsRune('a'), state.IsRune('A'),
		state.IsRune('b'), state.IsRune('B'),
		state.IsRune('c'), state.IsRune('C'),
		state.IsRune('d'), state.IsRune('D'),
		state.IsRune('e'), state.IsRune('E'),
		state.IsRune('f'), state.IsRune('F'),
	)

	// Number bodies
	binNumberBody = state.Or(
		state.IsRune('0'),
		state.IsRune('1'),
	)
	octNumberBody = state.Or(
		state.IsRune('0'),
		state.IsRune('1'),
		state.IsRune('2'),
		state.IsRune('3'),
		state.IsRune('4'),
		state.IsRune('5'),
		state.IsRune('6'),
		state.IsRune('7'),
	)
	decNumberBody = unicode.IsDigit
	hexNumberBody = isHexDigit

	// Number prefixes
	BinNumberPrefixes = []string{
		"0b", "0B",
	}
	OctNumberPrefixes = []string{
		"0o", "0O",
	}
	DecNumberPrefixes = []string{}
	HexNumberPrefixes = []string{
		"0x", "0X",
	}
)

func signedNumberGuard(ctx context.Context, _ xio.State) (err error) {
	provider, ok := state.GetHistoryProvider[Token](ctx)
	if !ok {
		// history is not enabled
		err = state.MakeErrBreak(ErrDisabledHistory)
		return
	}
	history := provider.Get()
	if len(history) == 0 {
		err = state.ErrRollback
		return
	}
	_, ok = allValueTokens[history[len(history)-1].Token]
	if ok {
		err = state.ErrRollback
	}
	return
}

func numberSubState(
	withFraction bool,
	numberBody func(rune) bool,
	maxBodyLen uint,
	maxFractionLen uint,
	errInvalidNumber error,
) func(b state.Builder[Token]) []state.Update[Token] {
	return func(b state.Builder[Token]) (subState []state.Update[Token]) {
		// consume all digits
		digits := func(b state.Builder[Token]) *state.Chain[Token] {
			return b.RuneCheck(numberBody).
				Repeat(state.CountBetween(0, maxBodyLen-1))
		}
		if withFraction {
			fraction := func(b state.Builder[Token]) []state.Update[Token] {
				return state.AsSlice[state.Update[Token]](
					b.RuneCheck(IsNumberDot).
						State(b, numberSubState(false, numberBody, maxFractionLen, 0, errInvalidNumber)).Optional().
						Break(),
				)
			}
			subState = state.AsSlice[state.Update[Token]](
				digits(b).State(b, fraction).Optional().FollowedByRuneCheck(allTerms).Break(),
				// if followed by non known term, emit error
				digits(b).State(b, fraction).Optional().FollowedByNotRuneCheck(numberBody).Rest().Error(errInvalidNumber),
				// otherwise, break
				digits(b).State(b, fraction).Optional().Break(),
			)
		} else {
			subState = state.AsSlice[state.Update[Token]](
				digits(b).FollowedByRuneCheck(allTerms).Break(),
				// if followed by non known term, emit error
				digits(b).FollowedByNotRuneCheck(numberBody).Rest().Error(errInvalidNumber),
				// otherwise, break
				digits(b).Break(),
			)
		}
		return
	}
}

func numberState(
	namePfx string,
	signed bool,
	withFraction bool,
	firstDigit func(rune) bool,
	numberBody func(rune) bool,
	requiredPrefixes []string,
	maxBodyLen uint,
	maxFractionLen uint,
	token Token,
	errInvalidNumber error,
) (number func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) (state *state.Chain[Token]) {
		state = b.Named(namePfx + token.String())
		if signed {
			state = state.RuneCheck(plusMinus).Optional().Tap(signedNumberGuard)
		}
		if len(requiredPrefixes) > 0 {
			state = state.String(requiredPrefixes...)
		}
		state = state.RuneCheck(firstDigit).
			State(b, numberSubState(withFraction, numberBody, maxBodyLen, maxFractionLen, errInvalidNumber)).Optional().
			Emit(token)
		return
	}
}

type numberStateBuilder struct {
	namePfx          string
	signed           bool
	withFraction     bool
	firstDigit       func(rune) bool
	numberBody       func(rune) bool
	requiredPrefixes []string
	maxBodyLen       uint
	maxFractionLen   uint
	token            Token
	errInvalidNumber error
}

func (nsb numberStateBuilder) build(b state.Builder[Token]) *state.Chain[Token] {
	return numberState(
		nsb.namePfx,
		nsb.signed,
		nsb.withFraction,
		nsb.firstDigit,
		nsb.numberBody,
		nsb.requiredPrefixes,
		nsb.maxBodyLen,
		nsb.maxFractionLen,
		nsb.token,
		nsb.errInvalidNumber,
	)(b)
}

var numberStateBuilders = []numberStateBuilder{
	// Bin fractions
	{"Signed", true, false, IsNumberDot, binNumberBody, BinNumberPrefixes, math.MaxUint, math.MaxUint, BinFraction, ErrInvalidNumber},
	{"Unsigned", false, false, IsNumberDot, binNumberBody, BinNumberPrefixes, math.MaxUint, math.MaxUint, BinFraction, ErrInvalidNumber},
	// Oct fractions
	{"Signed", true, false, IsNumberDot, octNumberBody, OctNumberPrefixes, math.MaxUint, math.MaxUint, OctFraction, ErrInvalidNumber},
	{"Unsigned", false, false, IsNumberDot, octNumberBody, OctNumberPrefixes, math.MaxUint, math.MaxUint, OctFraction, ErrInvalidNumber},
	// Hex fractions
	{"Signed", true, false, IsNumberDot, hexNumberBody, HexNumberPrefixes, math.MaxUint, math.MaxUint, HexFraction, ErrInvalidNumber},
	{"Unsigned", false, false, IsNumberDot, hexNumberBody, HexNumberPrefixes, math.MaxUint, math.MaxUint, HexFraction, ErrInvalidNumber},
	// Dec fractions
	{"Signed", true, false, IsNumberDot, decNumberBody, DecNumberPrefixes, math.MaxUint, math.MaxUint, DecFraction, ErrInvalidNumber},
	{"Unsigned", false, false, IsNumberDot, decNumberBody, DecNumberPrefixes, math.MaxUint, math.MaxUint, DecFraction, ErrInvalidNumber},

	// Bin numbers
	{"Signed", true, true, binNumberBody, binNumberBody, BinNumberPrefixes, math.MaxUint, math.MaxUint, BinNumber, ErrInvalidNumber},
	{"Unsigned", false, true, binNumberBody, binNumberBody, BinNumberPrefixes, math.MaxUint, math.MaxUint, BinNumber, ErrInvalidNumber},
	// Oct numbers
	{"Signed", true, true, octNumberBody, octNumberBody, OctNumberPrefixes, math.MaxUint, math.MaxUint, OctNumber, ErrInvalidNumber},
	{"Unsigned", false, true, octNumberBody, octNumberBody, OctNumberPrefixes, math.MaxUint, math.MaxUint, OctNumber, ErrInvalidNumber},
	// Hex numbers
	{"Signed", true, true, hexNumberBody, hexNumberBody, HexNumberPrefixes, math.MaxUint, math.MaxUint, HexNumber, ErrInvalidNumber},
	{"Unsigned", false, true, hexNumberBody, hexNumberBody, HexNumberPrefixes, math.MaxUint, math.MaxUint, HexNumber, ErrInvalidNumber},
	// Dec numbers
	{"Signed", true, true, decNumberBody, decNumberBody, DecNumberPrefixes, math.MaxUint, math.MaxUint, DecNumber, ErrInvalidNumber},
	{"Unsigned", false, true, decNumberBody, decNumberBody, DecNumberPrefixes, math.MaxUint, math.MaxUint, DecNumber, ErrInvalidNumber},
}
