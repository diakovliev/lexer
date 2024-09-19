package states

import (
	"os"
	"testing"
	"unicode"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/logger"
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

	factory := Make(logger, common.Dispose[testMessageType])

	chain1 := factory.Fn(unicode.IsDigit).Fn(unicode.IsDigit).Fn(unicode.IsDigit).Emit(common.User, messageType1)
	assert.NotNil(t, chain1)

	chain2 := factory.Fn(unicode.IsDigit).Fn(unicode.IsDigit).Fn(unicode.IsDigit).Emit(common.User, messageType1)
	assert.NotNil(t, chain2)
}
