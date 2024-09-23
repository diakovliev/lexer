package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Tap[T any] struct {
		logger common.Logger
		fn     TapFn
	}

	TapFn func(context.Context, xio.State) error
)

func newTap[T any](logger common.Logger, fn TapFn) *Tap[T] {
	return &Tap[T]{
		logger: logger,
		fn:     fn,
	}
}

func (t Tap[T]) Update(ctx context.Context, tx xio.State) (err error) {
	if err = t.fn(ctx, tx); err != nil {
		return
	}
	// return ErrCommit to be able to Tap end of the chain
	err = errCommit
	return
}

func (b Builder[T]) Tap(fn TapFn) (tail *Chain[T]) {
	defaultName := "Tap"
	tail = b.createNode(defaultName, func() any { return newTap[T](b.logger, fn) })
	return
}

func isTap[T any](s Update[T]) bool {
	_, ok := s.(*Tap[T])
	return ok
}
