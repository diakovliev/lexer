package lexer

import (
	"context"
	"io"
)

type (
	// Callback is a token callback
	Callback[T any] func(context.Context, Token[T]) error

	// AcceptR accepts a runes
	AcceptR func(ctx context.Context, next NextR) error

	// AcceptB accepts a bytes
	AcceptB func(ctx context.Context, next NextB) error

	// NextR is the next rune
	NextR interface {
		Next(ctx context.Context) (rune, error)
	}

	// NextB is the next byte
	NextB interface {
		Next(ctx context.Context) (byte, error)
	}

	Lexer[T any] interface {
		Run(ctx context.Context, reader io.Reader) error
		AcceptR(accept AcceptR) error
		AcceptB(accept AcceptB) error
		Emit(ctx context.Context, token T) error
	}

	LexerImpl[C Characher, T any] struct {
		options Options[C]
	}
)

func New[C Characher, T any](opts ...Option[C]) (ret *LexerImpl[C, T]) {
	ret = &LexerImpl[C, T]{}
	for _, opt := range opts {
		opt(&ret.options)
	}
	return
}

func (l *LexerImpl[C, T]) Run(ctx context.Context, reader io.Reader) (err error) {

	return
}
