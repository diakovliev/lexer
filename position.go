package lexer

// Position is the position of the token
type Position struct {
	// Start is the absolute position of the first character of the token
	Start int
	// End is the absolute position of the last character of the token
	End int

	// Line is the line of the token
	Line int
	// LineStart is the position of the first character of the token in the line
	LineStart int
	// LineEnd is the position of the last character of the token in the line
	LineEnd int
}
