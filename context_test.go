package lexer_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/diakovliev/lexer"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {

	type testCase struct {
		name      string
		input     string
		want      []lexer.Message[testMessageType]
		wantError error
	}

	tests := []testCase{
		{
			name:  "not empty string",
			input: `  -123 -baobab "this is \\ string \""`,
			want: []lexer.Message[testMessageType]{
				{Type: lexer.User, UserType: Number, Value: []byte("-123"), Pos: 2, Width: 4},
				{Type: lexer.User, UserType: Minus, Value: []byte("-"), Pos: 7, Width: 1},
				{Type: lexer.User, UserType: Identifier, Value: []byte("baobab"), Pos: 8, Width: 6},
				{Type: lexer.User, UserType: String, Value: []byte(`"this is \\ string \""`), Pos: 15, Width: 22},
			},
		},
		{
			name:  "empty string",
			input: `  -123 -baobab ""`,
			want: []lexer.Message[testMessageType]{
				{Type: lexer.User, UserType: Number, Value: []byte("-123"), Pos: 2, Width: 4},
				{Type: lexer.User, UserType: Minus, Value: []byte("-"), Pos: 7, Width: 1},
				{Type: lexer.User, UserType: Identifier, Value: []byte("baobab"), Pos: 8, Width: 6},
				{Type: lexer.User, UserType: String, Value: []byte(`""`), Pos: 15, Width: 2},
			},
		},
		{
			name:  "scope",
			input: `  -123 -baobab ( "1", "2" ) `,
			want: []lexer.Message[testMessageType]{
				{Type: lexer.User, UserType: Number, Value: []byte("-123"), Pos: 2, Width: 4},
				{Type: lexer.User, UserType: Minus, Value: []byte("-"), Pos: 7, Width: 1},
				{Type: lexer.User, UserType: Identifier, Value: []byte("baobab"), Pos: 8, Width: 6},
				{Type: lexer.User, UserType: Bra, Value: []byte("("), Pos: 15, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"1"`), Pos: 17, Width: 3},
				{Type: lexer.User, UserType: Comma, Value: []byte(","), Pos: 20, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"2"`), Pos: 22, Width: 3},
				{Type: lexer.User, UserType: Ket, Value: []byte(")"), Pos: 26, Width: 1},
			},
		},
		{
			name:  "scope 2",
			input: `  -123 -baobab ( "1", "2" ) 2345`,
			want: []lexer.Message[testMessageType]{
				{Type: lexer.User, UserType: Number, Value: []byte("-123"), Pos: 2, Width: 4},
				{Type: lexer.User, UserType: Minus, Value: []byte("-"), Pos: 7, Width: 1},
				{Type: lexer.User, UserType: Identifier, Value: []byte("baobab"), Pos: 8, Width: 6},
				{Type: lexer.User, UserType: Bra, Value: []byte("("), Pos: 15, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"1"`), Pos: 17, Width: 3},
				{Type: lexer.User, UserType: Comma, Value: []byte(","), Pos: 20, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"2"`), Pos: 22, Width: 3},
				{Type: lexer.User, UserType: Ket, Value: []byte(")"), Pos: 26, Width: 1},
				{Type: lexer.User, UserType: Number, Value: []byte("2345"), Pos: 28, Width: 4},
			},
		},
		{
			name:  "scope 3",
			input: `  -123 -baobab ( "1", "2", ("3", "4") ) 2345`,
			want: []lexer.Message[testMessageType]{
				{Type: lexer.User, UserType: Number, Value: []byte("-123"), Pos: 2, Width: 4},
				{Type: lexer.User, UserType: Minus, Value: []byte("-"), Pos: 7, Width: 1},
				{Type: lexer.User, UserType: Identifier, Value: []byte("baobab"), Pos: 8, Width: 6},
				{Type: lexer.User, UserType: Bra, Value: []byte("("), Pos: 15, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"1"`), Pos: 17, Width: 3},
				{Type: lexer.User, UserType: Comma, Value: []byte(","), Pos: 20, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"2"`), Pos: 22, Width: 3},
				{Type: lexer.User, UserType: Comma, Value: []byte(","), Pos: 25, Width: 1},
				{Type: lexer.User, UserType: Bra, Value: []byte("("), Pos: 27, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"3"`), Pos: 28, Width: 3},
				{Type: lexer.User, UserType: Comma, Value: []byte(","), Pos: 31, Width: 1},
				{Type: lexer.User, UserType: String, Value: []byte(`"4"`), Pos: 33, Width: 3},
				{Type: lexer.User, UserType: Ket, Value: []byte(")"), Pos: 36, Width: 1},
				{Type: lexer.User, UserType: Ket, Value: []byte(")"), Pos: 38, Width: 1},
				{Type: lexer.User, UserType: Number, Value: []byte("2345"), Pos: 40, Width: 4},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := lexer.NewTransactionReader(bytes.NewBufferString(tc.input))
			var messages []lexer.Message[testMessageType]
			yeild := func(msgs []lexer.Message[testMessageType]) {
				messages = append(messages, msgs...)
			}
			ctx := lexer.NewContext(reader.Begin(), yeild).Run(testInitialState)
			if tc.wantError != nil {
				assert.ErrorIs(t, ctx.Error, tc.wantError)
			} else {
				assert.ErrorIs(t, ctx.Error, io.EOF)
			}
			assert.Equal(t, tc.want, messages)
		})
	}
}
