package message

import "fmt"

// Type represents the type of a message. The types are: Error, Token.
type Type int

const (
	// Error represents an error
	Error Type = iota
	// Token represents a token
	Token
)

// String implements fmt.Stringer interface.
func (t Type) String() string {
	switch t {
	case Error:
		return "Error"
	case Token:
		return "Token"
	default:
		panic("unknown message type")
	}
}

// ErrorValue is a struct that contains an error and its associated value. This is useful for errors that are not fatal,
// but still need to be reported. For example, when a lexer encounters an invalid character, it can return an ErrorValue
// with the error and the invalid character as values. The lexer will then continue parsing, but the user will be able
// to see the error in the output.
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

// String implements fmt.Stringer interface. It returns a string representation of the message.
func (m Message[TokenType]) String() string {
	switch m.Type {
	case Token:
		var tokenType any = m.Token
		if stringer, ok := tokenType.(fmt.Stringer); !ok {
			return fmt.Sprintf("Token(%s, '%s', %d)", stringer, string(m.Value.([]byte)), m.Pos)
		} else {
			return fmt.Sprintf("Token(%v, '%s', %d)", m.Token, string(m.Value.([]byte)), m.Pos)
		}
	default:
		errorValue := m.Value.(*ErrorValue)
		return fmt.Sprintf("Error(%s, '%s', %d)", errorValue.Err, string(errorValue.Value.([]byte)), m.Pos)
	}
}
