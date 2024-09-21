package message

// Type represents the type of a message. The types are: Error, Token.
type Type int

const (
	// Error represents an error
	Error Type = iota
	// Token represents a token
	Token
)

type ErrorValue struct {
	Err   error
	Value any
}

// Message is the lexer's output type. It contains information about the lexeme and its type.
type Message[TokenType any] struct {
	// State level
	Level int
	// Type represents the message type. See MessageType for more details.
	Type Type
	// Token is only used when the message's type is Token. It contains the user-defined type of the lexeme.
	Token TokenType
	// Value represents the value of the lexeme.
	Value any
	// Pos
	Pos int
	// Width
	Width int
}
