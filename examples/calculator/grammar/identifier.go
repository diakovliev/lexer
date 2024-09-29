package grammar

import (
	"math"
	"unicode"

	"github.com/diakovliev/lexer/state"
)

var (
	identifierBody = state.Or(
		unicode.IsLetter,
		unicode.IsDigit,
		state.IsRune('_'),
	)
)

func identifierSubState(b state.Builder[Token]) []state.Update[Token] {
	body := func(b state.Builder[Token]) *state.Chain[Token] {
		return b.RuneCheck(identifierBody).Repeat(state.CountBetween(0, math.MaxUint))
	}
	return state.AsSlice[state.Update[Token]](
		body(b).FollowedByRuneCheck(allTerms).Break(),
		// if followed by non known term, emit error
		body(b).FollowedByNotRuneCheck(unicode.IsDigit).Rest().Error(ErrInvalidIdentifier),
		// otherwise, break
		body(b).Break(),
	)
}

func identifierState(name string) (identifier func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		return b.Named(name).
			// first letter must be a letter or underscore
			RuneCheck(unicode.IsLetter).
			// consume all subsequent letters, digits and underscores
			State(b, identifierSubState).
			// if followed by a known term, emit error
			Emit(Identifier)
	}
}
