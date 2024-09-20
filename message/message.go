package message

// MessageType represents the type of a message. The types are: Error, Token.
type MessageType int

const (
	// Error represents an error
	Error MessageType = iota
	// Token represents a token
	Token
)

// Message is the lexer's output type. It contains information about the lexeme and its type.
type Message[TokenType any] struct {
	// State level
	Level int
	// Type represents the message type. See MessageType for more details.
	Type MessageType
	// Token is only used when the message's type is Token. It contains the user-defined type of the lexeme.
	Token TokenType
	// Value represents the value of the lexeme.
	Value any
	// Pos
	Pos int
	// Width
	Width int
}
