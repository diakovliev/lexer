package lexer

import (
	"context"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Lexer is a lexical analyzer that reads input data and produces tokens.
	Lexer[T any] struct {
		logger   common.Logger
		source   xio.Source
		builder  state.Builder[T]
		provider state.Provider[T]
		messages []message.Message[T]
	}
)

// New creates a new lexer instance with the given reader and logger.
func New[T any](logger common.Logger, reader io.Reader) (ret *Lexer[T]) {
	ret = &Lexer[T]{
		logger: logger,
		source: xio.New(logger, reader),
	}
	ret.builder = state.Make(ret.logger, ret.receiver)
	return ret
}

// receiver is a function that receives messages from the lexer.
func (l *Lexer[T]) receiver(m message.Message[T]) (err error) {
	l.messages = append(l.messages, m)
	return nil
}

// Messages returns all messages produced by the lexer.
func (l Lexer[T]) Messages() []message.Message[T] {
	return l.messages
}

// With adds a new states produced by given provider to the lexer.
func (l *Lexer[T]) With(fn state.Provider[T]) *Lexer[T] {
	l.provider = fn
	return l
}

// Run runs the lexer until it is done or an error occurs.
func (l *Lexer[T]) Run(_ context.Context) (err error) {
	err = state.NewRun(l.logger, l.builder, l.provider, io.EOF).Run(l.source)
	return
}
