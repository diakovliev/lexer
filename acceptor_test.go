package lexer_test

import (
	"bytes"
	"io"
	"testing"
	"unicode"

	"github.com/diakovliev/lexer"
	"github.com/stretchr/testify/assert"
)

func TestAcceptor(t *testing.T) {
	type testCase struct {
		name      string
		input     string
		wantDone  bool
		wantEmit  bool
		want      lexer.Message[testMessageType]
		wantError error
		accept    func(ctx *lexer.Acceptor[testMessageType])
		assert    func(t *testing.T, reader *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType])
	}

	tests := []testCase{
		{
			name:  "Accept",
			input: "a+-b",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("a+-b"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.Accept(lexer.Rune('a')).
					Accept(lexer.Rune('+')).
					Accept(lexer.Rune('-')).
					Accept(lexer.Rune('b')).
					Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:     "Skip",
			input:    "a+-b",
			wantDone: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.Accept(lexer.Rune('a')).
					Accept(lexer.Rune('+')).
					Accept(lexer.Rune('-')).
					Accept(lexer.Rune('b')).
					Skip()
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:      "EOF",
			input:     "a+-",
			wantError: io.EOF,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.Accept(lexer.Rune('a')).
					Accept(lexer.Rune('+')).
					Accept(lexer.Rune('-')).
					Accept(lexer.Rune('b')).
					Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 0)
			},
		},
		{
			name:  "!Accept",
			input: "a+-b",
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.Accept(lexer.Rune('a')).
					Accept(lexer.Rune('+')).
					Accept(lexer.Rune('*')).
					Accept(lexer.Rune('b')).
					Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 0)
			},
		},
		{
			name:  "AsseptWhile",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("this"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptWhile(unicode.IsLetter).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:  "AsseptUntil",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("this"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptUntil(unicode.IsSpace).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:  "AcceptCount",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("this"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptCount(4).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:  "AcceptString",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("this"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptString("this").Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:  "AcceptString no match",
			input: "this is a test",
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptString("this1").Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 0)
			},
		},
		{
			name:  "AcceptAnyStringFrom",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("this"),
				Pos:      0,
				Width:    4,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptAnyStringFrom(
					"test",
					"is",
					"a",
					"this",
				).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 4)
			},
		},
		{
			name:  "AcceptAnyStringFrom no matches",
			input: "this is a test",
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptAnyStringFrom(
					"test1",
					"is2",
					"a3",
					"this4",
				).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 0)
			},
		},
		{
			name:  "AcceptAnyFrom",
			input: "this is a test",
			want: lexer.Message[testMessageType]{
				Type:     lexer.User,
				UserType: Identifier,
				Value:    []byte("t"),
				Pos:      0,
				Width:    1,
			},
			wantDone: true,
			wantEmit: true,
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptAnyFrom(
					lexer.Rune('a'),
					lexer.Rune('b'),
					lexer.Rune('c'),
					lexer.Rune('t'),
				).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 1)
			},
		},
		{
			name:  "AcceptAnyFrom no matches",
			input: "this is a test",
			accept: func(ctx *lexer.Acceptor[testMessageType]) {
				ctx.AcceptAnyFrom(
					lexer.Rune('a'),
					lexer.Rune('b'),
					lexer.Rune('c'),
					lexer.Rune('d'),
				).Emit(lexer.User, Identifier)
			},
			assert: func(t *testing.T, tx *lexer.ReaderTransaction, ctx *lexer.Acceptor[testMessageType]) {
				assert.True(t, tx.Pos() == 0)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			calledEmit := false
			emit := func(message lexer.Message[testMessageType]) {
				assert.Equal(t, tc.want, message)
				calledEmit = true
			}
			reader := lexer.NewTransactionReader(
				bytes.NewBufferString(tc.input),
			).Begin()
			tx := reader.Begin()
			ctx := lexer.NewAcceptor(
				tx,
				emit,
			)
			assert.False(t, ctx.Done())
			tc.accept(ctx)
			if tc.assert != nil {
				tc.assert(t, tx, ctx)
			}
			if tc.wantDone {
				assert.True(t, ctx.Done())
			} else {
				assert.False(t, ctx.Done())
			}
			assert.Equal(t, tc.wantEmit, calledEmit)
			if tc.wantError != nil {
				assert.Error(t, ctx.Error)
			} else {
				assert.NoError(t, ctx.Error)
			}
		})
	}
}
