package lexer_test

import (
	"bytes"
	"context"
	"io"
	"math"
	"os"
	"testing"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name         string
	input        string
	state        state.Provider[Token]
	wantMessages []*message.Message[Token]
	wantError    error
}

func TestLexer(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	tests := []testCase{}
	tests = append(tests, numberTests()...)

	tests = append(tests, []testCase{
		// Strings
		{
			name:  "string without escape",
			input: `"hello"`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"hello"`), Pos: 0, Width: 7},
			},
			wantError: io.EOF,
		},
		{
			name:  "2 strings without escape",
			input: `"hello" "world"`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"hello"`), Pos: 0, Width: 7},
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"world"`), Pos: 8, Width: 7},
			},
			wantError: io.EOF,
		},
		{
			name:  "string with escape in the middle",
			input: `"hel\"lo"`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"hel\"lo"`), Pos: 0, Width: 9},
			},
			wantError: io.EOF,
		},
		{
			name:  "string with escape at start",
			input: `"\"hello"`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"\"hello"`), Pos: 0, Width: 9},
			},
			wantError: io.EOF,
		},
		{
			name:  "string with escape at end",
			input: `"hello\""`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"hello\""`), Pos: 0, Width: 9},
			},
			wantError: io.EOF,
		},
		{
			name:  "string with multiply escapes",
			input: `"\"hello\""`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"\"hello\""`), Pos: 0, Width: 11},
			},
			wantError: io.EOF,
		},
		{
			name:  "string with multiply escapes 2",
			input: `"\"hel\\lo\""`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"\"hel\\lo\""`), Pos: 0, Width: 13},
			},
			wantError: io.EOF,
		},
		{
			name:  "2 string with escape",
			input: `"\"hello" "world\""`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"\"hello"`), Pos: 0, Width: 9},
				{Level: 0, Type: message.Token, Token: String, Value: []byte(`"world\""`), Pos: 10, Width: 9},
			},
			wantError: io.EOF,
		},
		{
			name:  "not closed string",
			input: `"hel\"lo`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Error, Value: &message.ErrorValue{Err: ErrInvalidExpression, Value: []byte(`"hel\"lo`)}, Pos: 0, Width: 8},
			},
			wantError: ErrInvalidExpression,
		},
		{
			name:  "not closed string 2",
			input: `"hello\"`,
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Error, Value: &message.ErrorValue{Err: ErrInvalidExpression, Value: []byte(`"hello\"`)}, Pos: 0, Width: 8},
			},
			wantError: ErrInvalidExpression,
		},

		// Mixed
		{
			name:  "sub state",
			input: "123 (123, 333) 555",
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("333"), Pos: 10, Width: 3},
				{Level: 1, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 13, Width: 1},
				{Level: 0, Type: message.Token, Token: DecNumber, Value: []byte("555"), Pos: 15, Width: 3},
			},
			wantError: io.EOF,
		},
		{
			name:  "sub state incomplete",
			input: "123 (123, 333 ",
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("333"), Pos: 10, Width: 3},
				{Level: 1, Type: message.Error, Value: &message.ErrorValue{Err: ErrInvalidExpression, Value: []byte("(123, 333 ")}, Pos: 4, Width: 10},
			},
			wantError: ErrInvalidExpression,
		},
		{
			name:  "inner sub states",
			input: "123 (123, 333, (1, 3, 4), 345) 555 foo bar",
			state: testGrammar(true, math.MaxUint),
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 0, Width: 3},
				{Level: 0, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 4, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("123"), Pos: 5, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 8, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("333"), Pos: 10, Width: 3},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 13, Width: 1},
				{Level: 1, Type: message.Token, Token: Bra, Value: []byte("("), Pos: 15, Width: 1},
				{Level: 2, Type: message.Token, Token: DecNumber, Value: []byte("1"), Pos: 16, Width: 1},
				{Level: 2, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 17, Width: 1},
				{Level: 2, Type: message.Token, Token: DecNumber, Value: []byte("3"), Pos: 19, Width: 1},
				{Level: 2, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 20, Width: 1},
				{Level: 2, Type: message.Token, Token: DecNumber, Value: []byte("4"), Pos: 22, Width: 1},
				{Level: 2, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 23, Width: 1},
				{Level: 1, Type: message.Token, Token: Comma, Value: []byte(","), Pos: 24, Width: 1},
				{Level: 1, Type: message.Token, Token: DecNumber, Value: []byte("345"), Pos: 26, Width: 3},
				{Level: 1, Type: message.Token, Token: Ket, Value: []byte(")"), Pos: 29, Width: 1},
				{Level: 0, Type: message.Token, Token: DecNumber, Value: []byte("555"), Pos: 31, Width: 3},
				{Level: 0, Type: message.Token, Token: Identifier, Value: []byte("foo"), Pos: 35, Width: 3},
				{Level: 0, Type: message.Token, Token: Identifier, Value: []byte("bar"), Pos: 39, Width: 3},
			},
			wantError: io.EOF,
		},
	}...)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			receiver := message.Slice[Token]()
			l := lexer.New(
				logger,
				bytes.NewBufferString(tc.input),
				message.DefaultFactory[Token](),
				receiver,
				lexer.WithHistoryDepth[Token](1),
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
