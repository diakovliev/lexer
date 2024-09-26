package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Tap is a state that calls the given function on Update.
	Tap[T any] struct {
		logger common.Logger
		fn     TapFn
	}

	// TapFn is a function that will be called on Update.
	TapFn func(context.Context, xio.State) error
)

// newTap creates a new instance of Tap state.
func newTap[T any](logger common.Logger, fn TapFn) *Tap[T] {
	return &Tap[T]{
		logger: logger,
		fn:     fn,
	}
}

// Update implements Update interface. It calls the given function on Update.
func (t Tap[T]) Update(ctx context.Context, tx xio.State) (err error) {
	if err = t.fn(ctx, tx); err != nil {
		return
	}
	// return errCommit to be able to Tap end of the chain
	err = ErrCommit
	return
}

// Tap adds a tap state to the chain. It calls the given function on Update.
func (b Builder[T]) Tap(callback TapFn) (tail *Chain[T]) {
	common.AssertNotNil(callback, "invalid grammar: nil callback")
	tail = b.append("Tap", func() Update[T] { return newTap[T](b.logger, callback) })
	return
}

// isTap checks if the given state is a tap state. It returns true if it is, false otherwise.
func isTap[T any](s Update[T]) bool {
	_, ok := s.(*Tap[T])
	return ok
}
