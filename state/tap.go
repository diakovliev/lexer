package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Tap is a state that calls the given function on Update.
	Tap[T any] struct {
		logger   common.Logger
		fn       TapFn
		factory  message.Factory[T]
		receiver message.Receiver[T]
	}

	// TapFn is a function that will be called on Update.
	TapFn func(context.Context, xio.State) error
)

// newTap creates a new instance of Tap state.
func newTap[T any](logger common.Logger, fn TapFn, factory message.Factory[T]) *Tap[T] {
	return &Tap[T]{
		logger:  logger,
		fn:      fn,
		factory: factory,
	}
}

func (t *Tap[T]) setReceiver(receiver message.Receiver[T]) {
	t.receiver = receiver
}

// Update implements Update interface. It calls the given function on Update.
func (t Tap[T]) Update(ctx context.Context, tx xio.State) (err error) {
	ctx = withFactory(ctx, t.factory)
	ctx = withReceiver(ctx, t.receiver)
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
	tap := newTap[T](b.logger, callback, b.factory)
	tail = b.append("Tap", func() Update[T] { return tap })
	// sent all messages to the the first node receiver
	tap.setReceiver(tail.head().receiver)
	return
}

// isTap checks if the given state is a tap state. It returns true if it is, false otherwise.
func isTap[T any](s Update[T]) bool {
	_, ok := s.(*Tap[T])
	return ok
}
