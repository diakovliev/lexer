package lexer

// lexeme is a lexeme
type lexeme struct {
	// s is a lexeme start position
	s int
	// w is a lexeme width
	w int
}

// pos returns the position of the Lexeme.
//
// It returns an integer representing the sum of the start and width fields of the Lexeme.
func (lm lexeme) pos() int {
	return lm.s + lm.w
}

// reset resets the width of the Lexeme.
//
// This function does not take any parameters and does not return anything.
func (lm *lexeme) reset() {
	lm.w = 0
}

// restore returns a function that restores the width of a Lexeme.
//
// It takes no parameters and returns a function that, when called, sets the
// width of the Lexeme to its original value.
func (lm *lexeme) restore() func() {
	width := lm.w
	return func() {
		lm.w = width
	}
}

// add increases the width of the lexeme by the specified delta.
//
// widthDelta - the amount by which to increase the width.
func (lm *lexeme) add(widthDelta int) {
	lm.w += widthDelta
}

// width returns the width of the Lexeme.
//
// This function does not take any parameters.
// It returns an integer representing the width of the Lexeme.
func (lm lexeme) width() int {
	return lm.w
}

// start returns the start value of the Lexeme.
//
// No parameters.
// Returns an integer.
func (lm lexeme) start() int {
	return lm.s
}

// next initializes a next Lexeme.
//
// It sets the start position and width of the Lexeme to 0,
// and returns a pointer to the next Lexeme.
func (lm lexeme) next() *lexeme {
	lm.s = lm.pos() //nolint:revive
	lm.w = 0        //nolint:revive
	return &lm
}

// from creates a new Lexeme instance with the given start position.
//
// start: the start position of the Lexeme
// returns: a pointer to the new Lexeme instance
func (lm lexeme) from(start int) *lexeme {
	lm.s = start //nolint:revive
	lm.w = 0     //nolint:revive
	return &lm
}
