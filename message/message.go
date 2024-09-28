package message

import (
	"fmt"

	"github.com/diakovliev/lexer/common"
)

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

// Error implements error interface.
func (ev ErrorValue) Error() string {
	if ev.Value == nil {
		return ev.Err.Error()
	}
	bytes, ok := ev.Value.([]byte)
	if ok {
		return fmt.Sprintf("%s: '%s'", ev.Err, string(bytes))
	}
	return fmt.Sprintf("%s: %#v", ev.Err, ev.Value)
}

// Unwrap implements error interface.
func (ev ErrorValue) Unwrap() error {
	return ev.Err
}

// Message is the lexer's output type. It contains information about the lexeme and its type.
type Message[TokenType any] struct {
	// State level. It can be useful if you want to build full AST tree from the messages.
	Level int
	// Type represents the message type. See MessageType for more details.
	Type Type
	// Token is only used when the message's type is Token. It contains the user-defined type of the lexeme.
	Token TokenType
	// Value represents the value of the lexeme. If you are using
	// default factory implementation, then this value will be either an error or a []byte in
	// dependence on the message's type.
	Value any
	// Pos is the position of the lexeme in the input.
	Pos int
	// Width is the width of the lexeme.
	Width int
}

// String implements fmt.Stringer interface. It returns a string representation of the message.
func (m Message[TokenType]) String() (ret string) {
	switch m.Type {
	case Token:
		var tokenType any = m.Token
		if stringer, ok := tokenType.(fmt.Stringer); ok {
			ret = fmt.Sprintf("Token(%s, %d, %d)", stringer, m.Pos, m.Width)
		} else {
			ret = fmt.Sprintf("Token(%v, %d, %d)", m.Token, m.Pos, m.Width)
		}
	case Error:
		if err, ok := m.Value.(*ErrorValue); ok {
			ret = fmt.Sprintf("Error(%s, %d, %d)", err, m.Pos, m.Width)
		} else {
			ret = fmt.Sprintf("Error(%v, %d, %d)", m.Value, m.Pos, m.Width)
		}
	default:
		common.AssertUnreachable("invalid message type: %d", m.Type)
	}
	return
}

// AsError returns the value of the message as an ErrorValue. It panics if the message's type is not Error,
// or if the Value is not an ErrorValue.
func (m Message[TokenType]) AsError() (value *ErrorValue) {
	common.AssertTrue(m.Type == Error, "invalid message type: %s", m.Type)
	value, ok := m.Value.(*ErrorValue)
	common.AssertTrue(ok, "value is not an error: %v", m.Value)
	return
}

// AsBytes returns the value of the message as a []byte. It panics if the message's type is not Token,
// or if the Value is not []byte.
func (m Message[TokenType]) AsBytes() (value []byte) {
	common.AssertTrue(m.Type == Token, "invalid message type: %s", m.Type)
	value, ok := m.Value.([]byte)
	common.AssertTrue(ok, "value is not []byte: %v", m.Value)
	return
}

// AsString returns the value of the message as a string. It panics if the message's type is not Token,
// or if the Value is not a string or []byte.
func (m Message[TokenType]) AsString() (value string) {
	common.AssertTrue(m.Type == Token, "invalid message type: %s", m.Type)
	if str, ok := m.Value.(string); ok {
		value = str
		return
	}
	if bytes, ok := m.Value.([]byte); ok {
		value = string(bytes)
		return
	}
	common.AssertUnreachable("invalid value type: %T", m.Value)
	return
}
