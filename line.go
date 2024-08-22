package lexer

// line is a line
type line struct {
	// number
	n int
	// start position
	s int
}

// next creates a next Line with the given start value.
//
// It takes an integer parameter 'start', which specifies the start value for the next Line.
// It returns a pointer to the newly created Line.
func (ln line) next(start int) *line {
	ln.n++       //nolint:revive
	ln.s = start //nolint:revive
	return &ln
}

// number returns the number of the Line.
//
// It takes no parameters.
// It returns an integer.
func (ln line) number() int {
	return ln.n
}

// start returns the start position of the Line.
//
// No parameters.
// Returns an integer.
func (ln line) start() int {
	return ln.s
}
