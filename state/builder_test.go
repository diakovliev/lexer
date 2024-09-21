package state

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
	"github.com/stretchr/testify/assert"
)

type fakeState struct{}
type notAState struct{}

func (t fakeState) Update(_ context.Context, tx xio.State) error {
	return errors.New("state error")
}

func newNotAState() any {
	return &notAState{}
}

func newFakeState() any {
	return &fakeState{}
}

func TestBuilder_NotAState(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	b := Make(
		logger,
		message.DefaultFactory[Token](),
		message.Dispose[Token](),
	)
	assert.Panics(t, func() { b.createNode("s0", newNotAState) })
}

func TestBuilder(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	b := Make(
		logger,
		message.DefaultFactory[Token](),
		message.Dispose[Token](),
	)

	s0 := b.createNode("s0", newFakeState)
	assert.NotNil(t, s0)
	assert.NotNil(t, s0.state)
	assert.Nil(t, s0.prev)
	assert.Nil(t, s0.next)
	assert.NotNil(t, s0.Builder.last)
	assert.Equal(t, s0, s0.Builder.last)
	assert.Equal(t, s0, s0.Tail())
	assert.Equal(t, s0, s0.Head())

	s1 := s0.createNode("s1", newFakeState)
	assert.NotNil(t, s1)
	assert.NotNil(t, s1.state)
	assert.Nil(t, s1.next)
	assert.Equal(t, s1.prev, s0)
	assert.Equal(t, s1, s0.next)
	assert.NotNil(t, s1.Builder.last)
	assert.Equal(t, s1, s1.Builder.last)
	assert.Equal(t, s1, s1.Tail())
	assert.Equal(t, s0, s1.Head())

	s2 := s1.createNode("s2", newFakeState)
	assert.NotNil(t, s2)
	assert.NotNil(t, s2.state)
	assert.Nil(t, s2.next)
	assert.Equal(t, s2.prev, s1)
	assert.Equal(t, s2, s1.next)
	assert.NotNil(t, s2.Builder.last)
	assert.Equal(t, s2, s2.Builder.last)
	assert.Equal(t, s2, s2.Tail())
	assert.Equal(t, s0, s2.Head())

	assert.Equal(t, s0, s1.Head())
	assert.Equal(t, s0, s2.Head())

	assert.Equal(t, s2, s0.Tail())
	assert.Equal(t, s2, s1.Tail())
}
