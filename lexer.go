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
		logger       common.Logger
		source       xio.Source
		builder      state.Builder[T]
		provider     state.Provider[T]
		historyDepth int
		history      message.History[T]
	}
)

// New creates a new lexer instance with the given reader and logger.
func New[T any](
	logger common.Logger,
	reader io.Reader,
	factory message.Factory[T],
	receiver message.Receiver[T],
	opts ...Option[T],
) (ret *Lexer[T]) {
	ret = &Lexer[T]{
		logger:       logger,
		source:       xio.New(logger, reader),
		historyDepth: 0,
	}
	for _, opt := range opts {
		opt(ret)
	}
	if ret.historyDepth > 0 {
		ret.history = message.Remember(receiver, ret.historyDepth)
		ret.builder = state.Make(
			logger,
			factory,
			ret.history,
		)
	} else {
		ret.builder = state.Make(
			logger,
			factory,
			receiver,
		)
	}
	return ret
}

// With adds a new states produced by given provider to the lexer.
func (l *Lexer[T]) With(fn state.Provider[T]) *Lexer[T] {
	l.provider = fn
	return l
}

// Run runs the lexer until it is done or an error occurs.
func (l *Lexer[T]) Run(ctx context.Context) (err error) {
	if l.history != nil {
		ctx = state.WithHistoryProvider(ctx, l.history)
	}
	err = state.NewRun(l.logger, l.builder, l.provider, io.EOF).
		Run(ctx, l.source)
	return
}
