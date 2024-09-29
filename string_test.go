package lexer_test

import "github.com/diakovliev/lexer/state"

// stringState matches string data between borders with appretiation of the escape character
func stringState(name string, escape rune, border rune) (identifier func(b state.Builder[Token]) *state.Chain[Token]) {
	return func(b state.Builder[Token]) *state.Chain[Token] {
		return b.Named(name).
			// Match the start of string
			Rune(border).
			// Consume string data
			UntilRune(state.Escape(state.IsRune(escape), state.IsRune(border)).Accept).
			// Match the end of string
			Rune(border).
			// We're done!
			Emit(String)

	}
}
