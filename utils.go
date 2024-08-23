package lexer

// Rune returns a function that checks if a given rune is equal to the input rune.
//
// It takes a parameter `ir` of type `rune`.
// It returns a function of type `func(r rune) bool`.
func Rune(ir rune) func(r rune) bool {
	return func(r rune) bool {
		return r == ir
	}
}

// Runes returns a function that checks if a given rune is present in the input string.
//
// The function takes a string as input and returns a function that takes a rune as input and returns a boolean value.
// The returned function checks if the given rune is present in the input string and returns true if it is, and false otherwise.
func Runes(input string) func(r rune) bool {
	return func(r rune) bool {
		for _, ir := range input {
			if r == ir {
				return true
			}
		}
		return false
	}
}

type EscapeCondition struct {
	escape  func(r rune) bool
	cond    func(r rune) bool
	escaped bool
}

func Escape(escape func(r rune) bool, cond func(r rune) bool) *EscapeCondition {
	return &EscapeCondition{
		escape:  escape,
		cond:    cond,
		escaped: false,
	}
}

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
		ret = true
	case isEscape && !e.escaped:
		// first escape symbol
		e.escaped = true
		ret = true
	case isEscape && e.escaped:
		// escape symbol followed by escape symbol
		e.escaped = false
		ret = true
	}
	return
}
