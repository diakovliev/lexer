package lexer

// Token is a token
// T is the type of the token
type Token[T any] struct {
	Position
	// Value is the value of the token
	Value T
	// Data is the data of the token
	Data string
}
