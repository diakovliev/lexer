package state

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
	"github.com/stretchr/testify/assert"
)

func TestRepeat_invalidQuantifier(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)
	receiver := message.Slice[testMessageType]()
	b := Make(logger, receiver)

	// min == max
	assert.NotNil(t, b.String("foo").Repeat(CountBetween(100, 100)).Emit(messageType1))
	// min == max = 0
	assert.NotNil(t, b.String("foo").Repeat(Count(0)).Emit(messageType1))
	// min == max = 1
	assert.NotNil(t, b.String("foo").Repeat(Count(1)).Emit(messageType1))
	// min > max
	assert.Panics(t, func() {
		b.String("foo").Repeat(CountBetween(100, 50)).Emit(messageType1)
	})
}

func TestRepeat(t *testing.T) {
	type testCase struct {
		name         string
		input        string
		state        func(b Builder[testMessageType]) *Chain[testMessageType]
		wantMessages []message.Message[testMessageType]
		wantError    error
	}

	tests := []testCase{
		{
			name:  `foofoofoo 'foo'.Count(3)`,
			input: "foofoofoo",
			state: func(b Builder[testMessageType]) *Chain[testMessageType] {
				return b.String("foo").Repeat(Count(3)).Emit(messageType1)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: messageType1, Value: []byte("foofoofoo"), Pos: 0, Width: 9},
			},
			wantError: ErrCommit,
		},
		{
			name:  "foofoofo 'foo'.count(3)",
			input: "foofoofo",
			state: func(b Builder[testMessageType]) *Chain[testMessageType] {
				return b.String("foo").Repeat(Count(3)).Emit(messageType1)
			},
			wantError: ErrRollback,
		},
		{
			name:  `foofoofo 'foo'.Count(2)`,
			input: "foofoofo",
			state: func(b Builder[testMessageType]) *Chain[testMessageType] {
				return b.String("foo").Repeat(Count(2)).Emit(messageType1)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: messageType1, Value: []byte("foofoo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `foofoofo 'foo'.CountBetween(2,3)`,
			input: "foofoofo",
			state: func(b Builder[testMessageType]) *Chain[testMessageType] {
				return b.String("foo").Repeat(CountBetween(2, 3)).Emit(messageType1)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: messageType1, Value: []byte("foofoo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `fooffo 'foo'.CountBetween(1,3)`,
			input: "fooffo",
			state: func(b Builder[testMessageType]) *Chain[testMessageType] {
				return b.String("foo").Repeat(CountBetween(1, 3)).Emit(messageType1)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: messageType1, Value: []byte("foo"), Pos: 0, Width: 3},
			},
			wantError: ErrCommit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := logger.New(
				logger.WithLevel(logger.Trace),
				logger.WithWriter(os.Stdout),
			)
			receiver := message.Slice[testMessageType]()
			b := Make(logger, receiver)
			tx := xio.New(logger, bytes.NewBufferString(tc.input))
			err := tc.state(b).Update(WithNextStateLevel(context.Background()), tx.Begin())
			assert.ErrorIs(t, err, tc.wantError)
			assert.Equal(t, tc.wantMessages, receiver.Slice)
		})
	}
}
