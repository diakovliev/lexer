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

func TestBuilder_append(t *testing.T) {
	b := makeTestDisposeBuilder()

	s0 := b.append("s0", newFakeState)
	assert.NotNil(t, s0)
	assert.NotNil(t, s0.ref)
	assert.NotNil(t, s0.receiver)
	assert.Nil(t, s0.p)
	assert.Nil(t, s0.n)
	assert.NotNil(t, s0.Builder.last)
	assert.Equal(t, s0, s0.Builder.last)
	assert.Equal(t, s0, s0.tail())
	assert.Equal(t, s0, s0.head())
	assert.Equal(t, "s0", s0.nodeName)

	s1 := s0.append("s1", newFakeState)
	assert.NotNil(t, s1)
	assert.NotNil(t, s1.ref)
	assert.Nil(t, s1.receiver)
	assert.Nil(t, s1.n)
	assert.Equal(t, s1.p, s0)
	assert.Equal(t, s1, s0.n)
	assert.NotNil(t, s1.Builder.last)
	assert.Equal(t, s1, s1.Builder.last)
	assert.Equal(t, s1, s1.tail())
	assert.Equal(t, s0, s1.head())
	assert.Equal(t, "s0.s1", s1.nodeName)

	s2 := s1.append("s2", newFakeState)
	assert.NotNil(t, s2)
	assert.NotNil(t, s2.ref)
	assert.Nil(t, s2.receiver)
	assert.Nil(t, s2.n)
	assert.Equal(t, s2.p, s1)
	assert.Equal(t, s2, s1.n)
	assert.NotNil(t, s2.Builder.last)
	assert.Equal(t, s2, s2.Builder.last)
	assert.Equal(t, s2, s2.tail())
	assert.Equal(t, s0, s2.head())
	assert.Equal(t, "s0.s1.s2", s2.nodeName)

	assert.Equal(t, s0, s1.head())
	assert.Equal(t, s0, s2.head())

	assert.Equal(t, s2, s0.tail())
	assert.Equal(t, s2, s1.tail())

	assert.Panics(t, func() {
		s1.append("s3", newFakeState)
	})
}
