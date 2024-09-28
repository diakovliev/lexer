package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

// Emit is a state what emits message.
type Emit[T any] struct {
	logger   common.Logger
	fn       func() T
	factory  message.Factory[T]
	receiver message.Receiver[T]
}

// newEmit creates new instance of Emit state.
func newEmit[T any](
	logger common.Logger,
	factory message.Factory[T],
	token func() T,
) *Emit[T] {
	return &Emit[T]{
		logger:  logger,
		factory: factory,
		fn:      token,
	}
}

func (e *Emit[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

// Update implements Update interface.
func (e Emit[T]) Update(ctx context.Context, tx xio.State) (err error) {
	common.AssertNotNil(e.receiver, "receiver is not set")
	data, pos, err := tx.Data()
	common.AssertNoError(err, "data error")
	common.AssertFalse(len(data) == 0, "nothing to emit")
	level, ok := GetTokenLevel(ctx)
	common.AssertTrue(ok, "no token level in context")
	msg, err := e.factory.Token(ctx, level, e.fn(), data, int(pos), len(data))
	common.AssertNoError(err, "messages factory error")
	err = e.receiver.Receive(msg)
	common.AssertNoError(err, "send message error")
	err = ErrCommit
	return
}

func (b Builder[T]) emitState(name string, token func() T) (tail *Chain[T]) {
	common.AssertNotNilPtr(b.last, "invalid grammar: emit can't be the first state in chain")
	newNode := newEmit(b.logger, b.factory, token)
	tail = b.append(name, func() Update[T] { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.head().receiver)
	return
}

// Emit emits given token.
func (b Builder[T]) Emit(token T) (tail *Chain[T]) {
	tail = b.emitState("Emit", func() T { return token })
	return
}

// EmitFn emits token received from the given function.
func (b Builder[T]) EmitFn(fn func() T) (tail *Chain[T]) {
	tail = b.emitState("EmitFn", fn)
	return
}

// isEmit checks if the state is Emit.
func isEmit[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Emit[T])
	return
}
