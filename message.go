package lexer

// MessageType represents the type of a lexeme. The types are: Error, Drop, EOF, NL, User.
type MessageType int

const (
	// Error represents an error
	Error MessageType = iota
	// Drop represents a dropped lexeme
	Drop
	// EOF represents an end of file
	EOF
	// NL represents a new line
	NL
	// User represents a user message
	User
)

// Message is the lexer's output type. It contains information about the lexeme and its type.
type Message[T any] struct {
	// Type represents the message type. See MessageType for more details.
	Type MessageType
	// UserType is only used when the message's type is User. It contains the user-defined type of the lexeme.
	UserType T
	// Value represents the value of the lexeme.
	Value []byte
}
