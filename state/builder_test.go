package state

import (
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/stretchr/testify/assert"
)

func makeTestDisposeBuilder() Builder[Token] {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	return Make(
		logger,
		message.DefaultFactory[Token](),
		message.Dispose[Token](),
	)
}

func makeTestBuilder[Token any](receiver message.Receiver[Token]) Builder[Token] {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	return Make(
		logger,
		message.DefaultFactory[Token](),
		receiver,
	)
}

func TestBuilder_NotAState(t *testing.T) {
	assert.Panics(t, func() { makeTestDisposeBuilder().append("s0", newNotAState) })
}

func TestBuilder_append(t *testing.T) {
	b := makeTestDisposeBuilder()

	s0 := b.append("s0", newFakeState)
	assert.NotNil(t, s0)
	assert.NotNil(t, s0.state)
	assert.Nil(t, s0.prev)
	assert.Nil(t, s0.next)
	assert.NotNil(t, s0.Builder.last)
	assert.Equal(t, s0, s0.Builder.last)
	assert.Equal(t, s0, s0.Tail())
	assert.Equal(t, s0, s0.Head())
	assert.Equal(t, "s0", s0.name)

	s1 := s0.append("s1", newFakeState)
	assert.NotNil(t, s1)
	assert.NotNil(t, s1.state)
	assert.Nil(t, s1.next)
	assert.Equal(t, s1.prev, s0)
	assert.Equal(t, s1, s0.next)
	assert.NotNil(t, s1.Builder.last)
	assert.Equal(t, s1, s1.Builder.last)
	assert.Equal(t, s1, s1.Tail())
	assert.Equal(t, s0, s1.Head())
	assert.Equal(t, "s0.s1", s1.name)

	s2 := s1.append("s2", newFakeState)
	assert.NotNil(t, s2)
	assert.NotNil(t, s2.state)
	assert.Nil(t, s2.next)
	assert.Equal(t, s2.prev, s1)
	assert.Equal(t, s2, s1.next)
	assert.NotNil(t, s2.Builder.last)
	assert.Equal(t, s2, s2.Builder.last)
	assert.Equal(t, s2, s2.Tail())
	assert.Equal(t, s0, s2.Head())
	assert.Equal(t, "s0.s1.s2", s2.name)

	assert.Equal(t, s0, s1.Head())
	assert.Equal(t, s0, s2.Head())

	assert.Equal(t, s2, s0.Tail())
	assert.Equal(t, s2, s1.Tail())

	// assert.Panics(t, func() {
	// 	s1.append("s3", newFakeState)
	// })
}
