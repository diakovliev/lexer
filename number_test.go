package lexer_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"unicode"

	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
	"github.com/diakovliev/lexer/xio"
)

var (
	allNumberTokens = map[Token]bool{
		BinNumber:   true,
		OctNumber:   true,
		DecNumber:   true,
		HexNumber:   true,
		BinFraction: true,
		OctFraction: true,
		DecFraction: true,
		HexFraction: true,
	}

	plusMinus = state.Or(
		state.IsRune('+'),
		state.IsRune('-'),
	)

	// Numbers misc
	isNumberDot = state.IsRune('.')
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
	IsBinDigit = state.Or(
		state.IsRune('0'),
		state.IsRune('1'),
	)
	IsOctDigit = state.Or(
		state.IsRune('0'),
		state.IsRune('1'),
		state.IsRune('2'),
		state.IsRune('3'),
		state.IsRune('4'),
		state.IsRune('5'),
		state.IsRune('6'),
		state.IsRune('7'),
	)
	IsDecDigit = unicode.IsDigit
	IsHexDigit = isHexDigit

	// Number prefixes
	BinPrefixes = []string{
		"0b", "0B",
	}
	OctPrefixes = []string{
		"0o", "0O",
	}
	DecPrefixes = []string{}
	HexPrefixes = []string{
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
		// In real application we probably want to reject number
		// here and generate operator. See calculator example.
		// err = state.ErrRollback
		return
	}
	_, ok = allNumberTokens[history[len(history)-1].Token]
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
					b.RuneCheck(isNumberDot).
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
	{"Signed", true, false, isNumberDot, IsBinDigit, BinPrefixes, math.MaxUint, math.MaxUint, BinFraction, ErrInvalidNumber},
	{"Unsigned", false, false, isNumberDot, IsBinDigit, BinPrefixes, math.MaxUint, math.MaxUint, BinFraction, ErrInvalidNumber},
	// Oct fractions
	{"Signed", true, false, isNumberDot, IsOctDigit, OctPrefixes, math.MaxUint, math.MaxUint, OctFraction, ErrInvalidNumber},
	{"Unsigned", false, false, isNumberDot, IsOctDigit, OctPrefixes, math.MaxUint, math.MaxUint, OctFraction, ErrInvalidNumber},
	// Hex fractions
	{"Signed", true, false, isNumberDot, IsHexDigit, HexPrefixes, math.MaxUint, math.MaxUint, HexFraction, ErrInvalidNumber},
	{"Unsigned", false, false, isNumberDot, IsHexDigit, HexPrefixes, math.MaxUint, math.MaxUint, HexFraction, ErrInvalidNumber},
	// Dec fractions
	{"Signed", true, false, isNumberDot, IsDecDigit, DecPrefixes, math.MaxUint, math.MaxUint, DecFraction, ErrInvalidNumber},
	{"Unsigned", false, false, isNumberDot, IsDecDigit, DecPrefixes, math.MaxUint, math.MaxUint, DecFraction, ErrInvalidNumber},

	// Bin numbers
	{"Signed", true, true, IsBinDigit, IsBinDigit, BinPrefixes, math.MaxUint, math.MaxUint, BinNumber, ErrInvalidNumber},
	{"Unsigned", false, true, IsBinDigit, IsBinDigit, BinPrefixes, math.MaxUint, math.MaxUint, BinNumber, ErrInvalidNumber},
	// Oct numbers
	{"Signed", true, true, IsOctDigit, IsOctDigit, OctPrefixes, math.MaxUint, math.MaxUint, OctNumber, ErrInvalidNumber},
	{"Unsigned", false, true, IsOctDigit, IsOctDigit, OctPrefixes, math.MaxUint, math.MaxUint, OctNumber, ErrInvalidNumber},
	// Hex numbers
	{"Signed", true, true, IsHexDigit, IsHexDigit, HexPrefixes, math.MaxUint, math.MaxUint, HexNumber, ErrInvalidNumber},
	{"Unsigned", false, true, IsHexDigit, IsHexDigit, HexPrefixes, math.MaxUint, math.MaxUint, HexNumber, ErrInvalidNumber},
	// Dec numbers
	{"Signed", true, true, IsDecDigit, IsDecDigit, DecPrefixes, math.MaxUint, math.MaxUint, DecNumber, ErrInvalidNumber},
	{"Unsigned", false, true, IsDecDigit, IsDecDigit, DecPrefixes, math.MaxUint, math.MaxUint, DecNumber, ErrInvalidNumber},
}

// /////////////////////////////////////////////////////////////////////////////
func randomNumberString(t Token, width int, capital bool) (ret string) {
	var n int
	for n == 0 {
		n = rand.IntN(math.MaxInt)
	}
	switch t {
	case DecNumber:
		ret += fmt.Sprintf("%d", n)
	case BinNumber:
		ret += fmt.Sprintf("%b", n)
	case OctNumber:
		ret += fmt.Sprintf("%o", n)
	case HexNumber:
		if capital {
			ret += fmt.Sprintf("%X", n)
		} else {
			ret += fmt.Sprintf("%x", n)
		}
	case DecFraction:
		ret += fmt.Sprintf("%d", n)
	case BinFraction:
		ret += fmt.Sprintf("%b", n)
	case OctFraction:
		ret += fmt.Sprintf("%o", n)
	case HexFraction:
		ret += fmt.Sprintf("%x", n)
	default:
		panic(fmt.Sprintf("unsupported token: %d", t))
	}
	if len(ret) < width {
		for len(ret) < width {
			ret += randomNumberString(t, width, capital)
		}
	}
	return ret[:width]
}

func randomNumberTest(wantLevel int, t Token, width int, capital bool, pfx, sfx string, wantError error) (ret testCase) {
	input := randomNumberString(t, width, capital)
	input = pfx + input + sfx
	totalWidth := len(input)
	testName := fmt.Sprintf("random number %s %d input %s", t, width, input)
	ret = testCase{
		name:      testName,
		input:     input,
		state:     testGrammar(true, math.MaxUint),
		wantError: wantError,
	}
	if errors.Is(ret.wantError, io.EOF) {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Token, Token: t, Value: []byte(input), Pos: 0, Width: totalWidth},
		}
	} else {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Error, Value: &message.ErrorValue{Err: wantError, Value: []byte(input)}, Pos: 0, Width: totalWidth},
		}
	}

	return
}

func randomNumberWithFractionTest(wantLevel int, t Token, width int, capital bool, pfx, sfx string, wantError error) (ret testCase) {
	number := randomNumberString(t, width, capital)
	fraction := randomNumberString(t, width, capital)
	input := pfx + number + "." + fraction + sfx
	totalWidth := len(input)
	testName := fmt.Sprintf("random number with fraction %s %d input %s", t, width, input)
	ret = testCase{
		name:      testName,
		input:     input,
		state:     testGrammar(true, math.MaxUint),
		wantError: wantError,
	}
	if errors.Is(ret.wantError, io.EOF) {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Token, Token: t, Value: []byte(input), Pos: 0, Width: totalWidth},
		}
	} else {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Error, Value: &message.ErrorValue{Err: wantError, Value: []byte(input)}, Pos: 0, Width: totalWidth},
		}
	}

	return
}

func randomFractionTest(wantLevel int, t Token, width int, capital bool, pfx, sfx string, wantError error) (ret testCase) {
	fraction := randomNumberString(t, width, capital)
	input := pfx + "." + fraction + sfx
	totalWidth := len(input)
	testName := fmt.Sprintf("random fraction %s %d input %s", t, width, input)
	ret = testCase{
		name:      testName,
		input:     input,
		state:     testGrammar(true, math.MaxUint),
		wantError: wantError,
	}
	if errors.Is(ret.wantError, io.EOF) {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Token, Token: t, Value: []byte(input), Pos: 0, Width: totalWidth},
		}
	} else {
		ret.wantMessages = []*message.Message[Token]{
			{Level: wantLevel, Type: message.Error, Value: &message.ErrorValue{Err: wantError, Value: []byte(input)}, Pos: 0, Width: totalWidth},
		}
	}

	return
}

func makeRange(start, end int) (ret []int) {
	// ret := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		ret = append(ret, i)
	}
	return ret
}

func numberTests() (tests []testCase) {

	inputWidth := makeRange(1, 16)

	// Numbers
	for _, width := range inputWidth {
		tests = append(tests, []testCase{
			randomNumberTest(0, BinNumber, width, false, BinPrefixes[0], "", io.EOF),
			randomNumberTest(0, BinNumber, width, false, BinPrefixes[1], "", io.EOF),
			randomNumberTest(0, BinNumber, width, false, BinPrefixes[0], ".", io.EOF),
			randomNumberTest(0, BinNumber, width, false, BinPrefixes[1], ".", io.EOF),
			randomNumberTest(0, BinNumber, width, false, "-"+BinPrefixes[0], "", io.EOF),
			randomNumberTest(0, BinNumber, width, false, "-"+BinPrefixes[1], "", io.EOF),
			randomNumberTest(0, BinNumber, width, false, "-"+BinPrefixes[0], ".", io.EOF),
			randomNumberTest(0, BinNumber, width, false, "-"+BinPrefixes[1], ".", io.EOF),

			randomNumberTest(0, OctNumber, width, false, OctPrefixes[0], "", io.EOF),
			randomNumberTest(0, OctNumber, width, false, OctPrefixes[1], "", io.EOF),
			randomNumberTest(0, OctNumber, width, false, OctPrefixes[0], ".", io.EOF),
			randomNumberTest(0, OctNumber, width, false, OctPrefixes[1], ".", io.EOF),
			randomNumberTest(0, OctNumber, width, false, "-"+OctPrefixes[0], "", io.EOF),
			randomNumberTest(0, OctNumber, width, false, "-"+OctPrefixes[1], "", io.EOF),
			randomNumberTest(0, OctNumber, width, false, "-"+OctPrefixes[0], ".", io.EOF),
			randomNumberTest(0, OctNumber, width, false, "-"+OctPrefixes[1], ".", io.EOF),

			randomNumberTest(0, DecNumber, width, false, "", "", io.EOF),
			randomNumberTest(0, DecNumber, width, false, "", ".", io.EOF),

			randomNumberTest(0, HexNumber, width, false, HexPrefixes[0], "", io.EOF),
			randomNumberTest(0, HexNumber, width, false, HexPrefixes[1], "", io.EOF),
			randomNumberTest(0, HexNumber, width, true, HexPrefixes[0], "", io.EOF),
			randomNumberTest(0, HexNumber, width, true, HexPrefixes[1], "", io.EOF),
			randomNumberTest(0, HexNumber, width, false, HexPrefixes[0], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, false, HexPrefixes[1], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, true, HexPrefixes[0], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, true, HexPrefixes[1], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, false, "-"+HexPrefixes[0], "", io.EOF),
			randomNumberTest(0, HexNumber, width, false, "-"+HexPrefixes[1], "", io.EOF),
			randomNumberTest(0, HexNumber, width, true, "-"+HexPrefixes[0], "", io.EOF),
			randomNumberTest(0, HexNumber, width, true, "-"+HexPrefixes[1], "", io.EOF),
			randomNumberTest(0, HexNumber, width, false, "-"+HexPrefixes[0], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, false, "-"+HexPrefixes[1], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, true, "-"+HexPrefixes[0], ".", io.EOF),
			randomNumberTest(0, HexNumber, width, true, "-"+HexPrefixes[1], ".", io.EOF),
		}...)
	}

	// Numbers with fractions
	for _, width := range inputWidth {
		tests = append(tests, []testCase{
			randomNumberWithFractionTest(0, BinNumber, width, false, BinPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, BinNumber, width, false, BinPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, BinNumber, width, false, BinPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, BinNumber, width, false, BinPrefixes[1], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(0, BinNumber, width, false, "-"+BinPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, BinNumber, width, false, "-"+BinPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, BinNumber, width, false, "-"+BinPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, BinNumber, width, false, "-"+BinPrefixes[1], ".", ErrInvalidNumber),

			randomNumberWithFractionTest(0, OctNumber, width, false, OctPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, OctNumber, width, false, OctPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, OctNumber, width, false, OctPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, OctNumber, width, false, OctPrefixes[1], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(0, OctNumber, width, false, "-"+OctPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, OctNumber, width, false, "-"+OctPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, OctNumber, width, false, "-"+OctPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, OctNumber, width, false, "-"+OctPrefixes[1], ".", ErrInvalidNumber),

			randomNumberWithFractionTest(0, DecNumber, width, false, "", "", io.EOF),
			randomNumberWithFractionTest(3, DecNumber, width, false, "", ".", ErrInvalidNumber),

			randomNumberWithFractionTest(0, HexNumber, width, false, HexPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, false, HexPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, true, HexPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, true, HexPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, HexNumber, width, false, HexPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, false, HexPrefixes[1], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, true, HexPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, true, HexPrefixes[1], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(0, HexNumber, width, false, "-"+HexPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, false, "-"+HexPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, true, "-"+HexPrefixes[0], "", io.EOF),
			randomNumberWithFractionTest(0, HexNumber, width, true, "-"+HexPrefixes[1], "", io.EOF),
			randomNumberWithFractionTest(3, HexNumber, width, false, "-"+HexPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, false, "-"+HexPrefixes[1], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, true, "-"+HexPrefixes[0], ".", ErrInvalidNumber),
			randomNumberWithFractionTest(3, HexNumber, width, true, "-"+HexPrefixes[1], ".", ErrInvalidNumber),
		}...)
	}

	// Fractions
	for _, width := range inputWidth {
		tests = append(tests, []testCase{
			randomFractionTest(0, BinFraction, width, false, BinPrefixes[0], "", io.EOF),
			randomFractionTest(0, BinFraction, width, false, BinPrefixes[1], "", io.EOF),
			randomFractionTest(1, BinFraction, width, false, BinPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, BinFraction, width, false, BinPrefixes[1], ".", ErrInvalidNumber),
			randomFractionTest(0, BinFraction, width, false, "-"+BinPrefixes[0], "", io.EOF),
			randomFractionTest(0, BinFraction, width, false, "-"+BinPrefixes[1], "", io.EOF),
			randomFractionTest(1, BinFraction, width, false, "-"+BinPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, BinFraction, width, false, "-"+BinPrefixes[1], ".", ErrInvalidNumber),

			randomFractionTest(0, OctFraction, width, false, OctPrefixes[0], "", io.EOF),
			randomFractionTest(0, OctFraction, width, false, OctPrefixes[1], "", io.EOF),
			randomFractionTest(1, OctFraction, width, false, OctPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, OctFraction, width, false, OctPrefixes[1], ".", ErrInvalidNumber),
			randomFractionTest(0, OctFraction, width, false, "-"+OctPrefixes[0], "", io.EOF),
			randomFractionTest(0, OctFraction, width, false, "-"+OctPrefixes[1], "", io.EOF),
			randomFractionTest(1, OctFraction, width, false, "-"+OctPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, OctFraction, width, false, "-"+OctPrefixes[1], ".", ErrInvalidNumber),

			randomFractionTest(0, DecFraction, width, false, "", "", io.EOF),
			randomFractionTest(1, DecFraction, width, false, "", ".", ErrInvalidNumber),
			// crazy cases "..NNN"
			randomFractionTest(1, DecFraction, width, false, ".", "", ErrInvalidNumber),
			// crazy cases "...NNN"
			randomFractionTest(1, DecFraction, width, false, "..", "", ErrInvalidNumber),

			randomFractionTest(0, HexFraction, width, false, HexPrefixes[0], "", io.EOF),
			randomFractionTest(0, HexFraction, width, false, HexPrefixes[1], "", io.EOF),
			randomFractionTest(0, HexFraction, width, true, HexPrefixes[0], "", io.EOF),
			randomFractionTest(0, HexFraction, width, true, HexPrefixes[1], "", io.EOF),
			randomFractionTest(1, HexFraction, width, false, HexPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, false, HexPrefixes[1], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, true, HexPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, true, HexPrefixes[1], ".", ErrInvalidNumber),
			randomFractionTest(0, HexFraction, width, false, "-"+HexPrefixes[0], "", io.EOF),
			randomFractionTest(0, HexFraction, width, false, "-"+HexPrefixes[1], "", io.EOF),
			randomFractionTest(0, HexFraction, width, true, "-"+HexPrefixes[0], "", io.EOF),
			randomFractionTest(0, HexFraction, width, true, "-"+HexPrefixes[1], "", io.EOF),
			randomFractionTest(1, HexFraction, width, false, "-"+HexPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, false, "-"+HexPrefixes[1], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, true, "-"+HexPrefixes[0], ".", ErrInvalidNumber),
			randomFractionTest(1, HexFraction, width, true, "-"+HexPrefixes[1], ".", ErrInvalidNumber),
		}...)
	}

	return
}
