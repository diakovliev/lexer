package state

// EscapeCondition is a condition that checks if the input rune is escaped by another rune.
// It is designed to be used in Until state to parse strings with escape characters.
// See: grammar_test.go: stringState for an example of using Escape in Until state.
type EscapeCondition struct {
	escape  func(r rune) bool
	cond    func(r rune) bool
	escaped bool
}

// Escape returns a function that checks if the input rune is escaped by another rune.
// It is designed to be used in While state to parse strings with escape characters.
// See: grammar_test.go: stringState for an example of using Escape in Until state.
func Escape(escape func(r rune) bool, cond func(r rune) bool) *EscapeCondition {
	return &EscapeCondition{
		escape:  escape,
		cond:    cond,
		escaped: false,
	}
}

// Accept is a predicate that checks if the input rune is escaped by another rune.
func (e *EscapeCondition) Accept(r rune) (ret bool) {
	isEscape := e.escape(r)
	isCond := e.cond(r)
	switch {
	case !isEscape && !e.escaped:
		e.escaped = false
		ret = isCond
	case !isEscape && e.escaped:
		// escaped symbol
		e.escaped = false
		ret = false
	case isEscape && !e.escaped:
		// first escape symbol
		e.escaped = true
		ret = false
	case isEscape && e.escaped:
		// escape symbol followed by escape symbol
		e.escaped = false
		ret = false
	default:
		ret = true
	}
	return
}
