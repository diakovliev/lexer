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
	err = ErrCommit
	return
}

func (b Builder[T]) Tap(ctx context.Context, fn TapFn) (head *Chain[T]) {
	defaultName := "Tap"
	head = b.createNode(defaultName, func() any { return newTap[T](b.logger, fn) })
	return
}
