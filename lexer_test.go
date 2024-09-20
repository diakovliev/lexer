package lexer_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"unicode"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
	"github.com/stretchr/testify/assert"
)

type testMessageType int

var errUnhandledData = errors.New("unhandled data")

const (
	Number testMessageType = iota
	Bra
	Ket
	Comma
	Term
)

func buildScopeState(b state.Builder[testMessageType]) []state.State[testMessageType] {
	return state.AsSlice[state.State[testMessageType]](
		b.While(unicode.IsSpace).Omit(),
		b.Rune('(').Emit(Bra).State(b, buildScopeState),
		b.Rune(',').Emit(Comma),
		b.Rune(')').Emit(Ket).Break(),
		b.While(unicode.IsDigit).Emit(Number),
		b.String("foo").Emit(Term),
		b.String("bar").Emit(Term),
		b.Rest().Error(errUnhandledData),
	)
}

func buildInitialState(b state.Builder[testMessageType]) []state.State[testMessageType] {
	return state.AsSlice[state.State[testMessageType]](
		b.While(unicode.IsSpace).Omit(),
		b.Rune('(').Emit(Bra).State(b, buildScopeState),
		b.While(unicode.IsDigit).Emit(Number),
		b.String("foo").Emit(Term),
		b.String("bar").Emit(Term),
		b.Rest().Error(errUnhandledData),
	)
}

func TestLexer(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	type testCase struct {
		name         string
		input        string
		state        state.Provider[testMessageType]
		wantMessages []message.Message[testMessageType]
		wantError    error
	}

	tests := []testCase{
		{
			name:  "simple accept-fn",
			input: "123",
			state: func(b state.Builder[testMessageType]) []state.State[testMessageType] {
				return state.AsSlice[state.State[testMessageType]](
					b.Fn(unicode.IsDigit).Fn(unicode.IsDigit).Fn(unicode.IsDigit).Emit(Number),
					b.Rest().Error(errUnhandledData),
				)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 0, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "simple accept-fn 1",
			input: "123 345",
			state: func(b state.Builder[testMessageType]) []state.State[testMessageType] {
				return state.AsSlice[state.State[testMessageType]](
					b.While(unicode.IsDigit).Emit(Number).
						While(unicode.IsSpace).Omit().
						While(unicode.IsDigit).Emit(Number),
					b.Rest().Error(errUnhandledData),
				)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("345"), Pos: 4, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "simple accept-fn with spaces",
			input: "  123  ",
			state: func(b state.Builder[testMessageType]) []state.State[testMessageType] {
				return state.AsSlice[state.State[testMessageType]](
					b.While(unicode.IsSpace).Omit(),
					b.While(unicode.IsDigit).Emit(Number),
					b.Rest().Error(errUnhandledData),
				)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 2, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "simple accept-fn with spaces",
			input: "  1  ",
			state: func(b state.Builder[testMessageType]) []state.State[testMessageType] {
				return state.AsSlice[state.State[testMessageType]](
					b.While(unicode.IsSpace).Omit(),
					b.While(unicode.IsDigit).Emit(Number),
					b.Rest().Error(errUnhandledData),
				)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("1"), Pos: 2, Width: 1},
			},
			wantError: io.EOF,
		},
		{
			name:  "unhandled data",
			input: "123",
			state: func(b state.Builder[testMessageType]) []state.State[testMessageType] {
				return state.AsSlice[state.State[testMessageType]](
					b.Rest().Error(errUnhandledData),
				)
			},
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Error, Value: &state.ErrorValue{Err: errUnhandledData, Value: []byte("123")}, Pos: 0, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "substate",
			input: "123 (123, 333) 555",
			state: buildInitialState,
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("333"), Pos: 10, Width: 3},
				{Level: 1, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 13, Width: 1},
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("555"), Pos: 15, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "substate incomplete",
			input: "123 (123, 333 ",
			state: buildInitialState,
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("333"), Pos: 10, Width: 3},
			},
			wantError: state.ErrIncompleteState,
		},
		{
			name:  "inner substates",
			input: "123 (123, 333, (1, 3, 4), 345) 555 foo bar",
			state: buildInitialState,
			wantMessages: []message.Message[testMessageType]{
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("333"), Pos: 10, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 13, Width: 1},
				{Level: 1, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 15, Width: 1},
				{Level: 2, Type: message.Token, Token: Number, Value: []byte("1"), Pos: 16, Width: 1},
				{Level: 2, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 17, Width: 1},
				{Level: 2, Type: message.Token, Token: Number, Value: []byte("3"), Pos: 19, Width: 1},
				{Level: 2, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 20, Width: 1},
				{Level: 2, Type: message.Token, Token: Number, Value: []byte("4"), Pos: 22, Width: 1},
				{Level: 2, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 23, Width: 1},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 24, Width: 1},
				{Level: 1, Type: message.Token, Token: Number, Value: []byte("345"), Pos: 26, Width: 3},
				{Level: 1, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 29, Width: 1},
				{Level: 0, Type: message.Token, Token: Number, Value: []byte("555"), Pos: 31, Width: 3},
				{Level: 0, Type: message.Token, Token: Term, Value: []byte("foo"), Pos: 35, Width: 3},
				{Level: 0, Type: message.Token, Token: Term, Value: []byte("bar"), Pos: 39, Width: 3},
			},
			wantError: io.EOF,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			receiver := message.Slice[testMessageType]()
			l := lexer.New(
				logger,
				bytes.NewBufferString(tc.input),
				receiver,
			).With(tc.state)
			err := l.Run(context.Background())
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantMessages, receiver.Slice)
		})
	}
}
