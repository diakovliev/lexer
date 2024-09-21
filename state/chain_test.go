package state

import (
	"os"
	"testing"
	"unicode"

	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/stretchr/testify/assert"
)

type testMessageType int

const (
	messageType1 testMessageType = iota
	messageType2
	messageType3
)

func TestNode(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	b := Make(logger, message.Dispose[testMessageType]())
	chain1 := b.Fn(unicode.IsDigit).Fn(unicode.IsDigit).Fn(unicode.IsDigit).Emit(messageType1)
	assert.NotNil(t, chain1)
	chain2 := b.Fn(unicode.IsDigit).Fn(unicode.IsDigit).Fn(unicode.IsDigit).Emit(messageType1)
	assert.NotNil(t, chain2)
}
