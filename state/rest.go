package state

import (
	"context"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Rest[T any] struct {
		logger common.Logger
	}
)

func newRest[T any](logger common.Logger) *Rest[T] {
	return &Rest[T]{
		logger: logger,
	}
}

func (r *Rest[T]) Update(ctx context.Context, tx xio.State) (err error) {
	// just advance the reader and do nothing else
	_, _ = io.Copy(io.Discard, tx)
	err = ErrNext
	return
}

func (b Builder[T]) Rest() (tail *Chain[T]) {
	defaultName := "Rest"
	tail = b.createNode(defaultName, func() any { return newRest[T](b.logger) })
	return
}

func isRest[T any](s Update[T]) bool {
	_, ok := s.(*Rest[T])
	return ok
}
