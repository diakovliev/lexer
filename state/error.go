package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

// Error is a state that produces an error message.
type Error[T any] struct {
	logger   common.Logger
	fn       func() error
	factory  message.Factory[T]
	receiver message.Receiver[T]
}

// newError creates a new instance of the Error state.
func newError[T any](logger common.Logger, factory message.Factory[T], fn func() error) *Error[T] {
	return &Error[T]{
		logger:  logger,
		factory: factory,
		fn:      fn,
	}
}

// setReceiver sets the receiver of the state.
func (e *Error[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

// Update implements the Update interface. It produces an error message.
func (e Error[T]) Update(ctx context.Context, tx xio.State) (err error) {
	common.AssertNotNil(e.receiver, "receiver is not set")
	data, pos, err := tx.Buffer()
	common.AssertNoError(err, "get buffer error")
	if len(data) == 0 {
		err = ErrRollback
		return
	}
	level, ok := GetTokenLevel(ctx)
	common.AssertTrue(ok, "no token level in context")
	msg, err := e.factory.Error(ctx, level, e.fn(), data, int(pos), len(data))
	if err != nil {
		err = MakeErrBreak(err)
		return
	}
	err = e.receiver.Receive(msg)
	if err != nil {
		err = MakeErrBreak(err)
		return
	}
	err = MakeErrBreak(msg.AsError())
	return
}

func (b Builder[T]) errorState(name string, fn func() error) (tail *Chain[T]) {
	common.AssertNotNilPtr(b.last, "invalid grammar: emit can't be the first state in chain")
	newNode := newError(b.logger, b.factory, fn)
	tail = b.append(name, func() Update[T] { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.head().receiver)
	return
}

// ErrorFn emits error received from the given function.
func (b Builder[T]) ErrorFn(fn func() error) (tail *Chain[T]) {
	common.AssertNotNil(fn, "invalid grammar: nil error")
	return b.errorState("Error", fn)
}

// Error emits given error.
func (b Builder[T]) Error(err error) (tail *Chain[T]) {
	common.AssertNotNil(err, "invalid grammar: nil error")
	return b.errorState("Error", func() error { return err })
}

// isError returns true if the state is Error.
func isError[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Error[T])
	return
}
