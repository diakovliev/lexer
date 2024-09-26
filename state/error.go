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
	err      error
	factory  message.Factory[T]
	receiver message.Receiver[T]
}

// newError creates a new instance of the Error state.
func newError[T any](logger common.Logger, factory message.Factory[T], err error) *Error[T] {
	return &Error[T]{
		logger:  logger,
		factory: factory,
		err:     err,
	}
}

// setReceiver sets the receiver of the state.
func (e *Error[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

// Update implements the Update interface. It produces an error message.
func (e Error[T]) Update(ctx context.Context, tx xio.State) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	data, pos, err := tx.Buffer()
	if err != nil {
		e.logger.Fatal("get buffer error: %s", err)
	}
	if len(data) == 0 {
		err = ErrRollback
		return
	}
	level, ok := GetTokenLevel(ctx)
	if !ok {
		e.logger.Fatal("no token level in context")
	}
	msg, err := e.factory.Error(ctx, level, e.err, data, int(pos), len(data))
	if err != nil {
		e.logger.Fatal("messages factory error: %s", err)
	}
	err = e.receiver.Receive(msg)
	if err != nil {
		e.logger.Fatal("receiver error: %s", err)
	}
	err = makeErrBreak(e.err)
	return
}

// Error adds Error state to the chain.
func (b Builder[T]) Error(err error) (tail *Chain[T]) {
	if b.last == nil {
		b.logger.Fatal("invalid grammar: error can't be the first state in chain")
	}
	newNode := newError(b.logger, b.factory, err)
	tail = b.append("Error", func() Update[T] { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.head().receiver)
	return
}

// isError returns true if the state is Error.
func isError[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Error[T])
	return
}
