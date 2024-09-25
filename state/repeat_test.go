package state

import (
	"bytes"
	"context"
	"errors"
	"math"
	"testing"
	"unicode"

	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
	"github.com/stretchr/testify/assert"
)

func TestRepeat_Builder(t *testing.T) {
	type testCase struct {
		name      string
		state     func(b Builder[Token]) *Chain[Token]
		wantPanic bool
	}

	tests := []testCase{
		{
			name: `min == max`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(CountBetween(100, 100)).Emit(Token1)
			},
		},
		{
			name: `min == max = 0`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(0)).Emit(Token1)
			},
		},
		{
			name: `min == max = 1`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(1)).Emit(Token1)
			},
		},
		{
			name: `min > max`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(CountBetween(100, 50)).Emit(Token1)
			},
			wantPanic: true,
		},
		{
			name: `can't repeat repeat`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(1)).Repeat(CountBetween(100, 50)).Emit(Token1)
			},
			wantPanic: true,
		},
		{
			name: `can't repeat emit`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Emit(Token1).Repeat(Count(1))
			},
			wantPanic: true,
		},
		{
			name: `can't repeat error`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Error(errors.New("test")).Repeat(Count(1))
			},
			wantPanic: true,
		},
		{
			name: `can't repeat omit`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Omit().Repeat(Count(1))
			},
			wantPanic: true,
		},
		{
			name: `can't repeat rest`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.Rest().Repeat(Count(1))
			},
			wantPanic: true,
		},
		{
			name: `can't repeat tap`,
			state: func(b Builder[Token]) *Chain[Token] {
				return b.Tap(func(context.Context, xio.State) error { return nil }).Repeat(Count(1))
			},
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			builder := makeTestDisposeBuilder()
			if tc.wantPanic {
				assert.Panics(t, func() {
					tc.state(builder)
				})
			} else {
				assert.NotNil(t, tc.state(builder))
			}
		})
	}
}

func TestRepeat(t *testing.T) {
	type testCase struct {
		name         string
		input        string
		state        func(b Builder[Token]) *Chain[Token]
		wantMessages []*message.Message[Token]
		wantError    error
	}

	tests := []testCase{
		{
			name:  `_1,_ isDigit.isDigit.CountBetween(0,math.MaxUint)`,
			input: "1,",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.CheckRune(unicode.IsDigit).CheckRune(unicode.IsDigit).Repeat(CountBetween(0, math.MaxUint)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("1"), Pos: 0, Width: 1},
			},
			wantError: ErrCommit,
		},
		{
			name:  `foofoofoo 'foo'.Count(3)`,
			input: "foofoofoo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(3)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foofoofoo"), Pos: 0, Width: 9},
			},
			wantError: ErrCommit,
		},
		{
			name:  "foofoofo 'foo'.Count(3)",
			input: "foofoofo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(3)).Emit(Token1)
			},
			wantError:    ErrRollback,
			wantMessages: []*message.Message[Token]{},
		},
		{
			name:  `foofoofo 'foo'.Count(2)`,
			input: "foofoofo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(Count(2)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foofoo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `foofoofo 'foo'.CountBetween(2,3)`,
			input: "foofoofo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(CountBetween(2, 3)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foofoo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `fooffo 'foo'.CountBetween(1,3)`,
			input: "fooffo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").Repeat(CountBetween(1, 3)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foo"), Pos: 0, Width: 3},
			},
			wantError: ErrCommit,
		},
		{
			name:  `fooffo 'foo'.'ffo'.CountBetween(0,3)`,
			input: "fooffo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").String("ffo").Repeat(CountBetween(0, 3)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("fooffo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `foo 'foo'.'ffo'.CountBetween(0,3)`,
			input: "foo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").String("ffo").Repeat(CountBetween(0, 3)).Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foo"), Pos: 0, Width: 3},
			},
			wantError: ErrCommit,
		},
		{
			name:  `fooffo 'foo'.'ffo'.Optional()`,
			input: "fooffo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").String("ffo").Optional().Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("fooffo"), Pos: 0, Width: 6},
			},
			wantError: ErrCommit,
		},
		{
			name:  `foo 'foo'.'ffo'.Optional()`,
			input: "foo",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.String("foo").String("ffo").Optional().Emit(Token1)
			},
			wantMessages: []*message.Message[Token]{
				{Level: 0, Type: message.Token, Token: Token1, Value: []byte("foo"), Pos: 0, Width: 3},
			},
			wantError: ErrCommit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			receiver := message.Slice[Token]()
			builder := makeTestBuilder(receiver)
			source := xio.New(builder.logger, bytes.NewBufferString(tc.input))
			err := tc.state(builder).Update(WithNextTokenLevel(context.Background()), source.Begin().Ref)
			assert.ErrorIs(t, err, tc.wantError)
			assert.Equal(t, tc.wantMessages, receiver.Slice)
		})
	}
}
