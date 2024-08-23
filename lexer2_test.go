package lexer_test

import (
	"bytes"
	"testing"

	"github.com/diakovliev/lexer"
	"github.com/stretchr/testify/assert"
)

type (
	testMessageType int
	testState       struct{}
)

const (
	Identifier testMessageType = iota
	Plus
	Minus
	Number
	String
	Bra
	Ket
	Comma
	Last
)

func (ts testState) State(lex *lexer.Lexer2[testMessageType]) lexer.StateFn2[testMessageType] {
	return nil
}

func TestLexer2(t *testing.T) {
	var messages []lexer.Message[testMessageType]
	l := lexer.New2(
		bytes.NewBufferString("a+-b"),
		&testState{},
	).WithCallback(func(m lexer.Message[testMessageType]) error {
		messages = append(messages, m)
		return nil
	})
	assert.NotNil(t, l)
}
