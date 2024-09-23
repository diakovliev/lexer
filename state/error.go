package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	Error[T any] struct {
		logger   common.Logger
		err      error
		factory  message.Factory[T]
		receiver message.Receiver[T]
	}
)

func newError[T any](
	logger common.Logger,
	factory message.Factory[T],
	err error,
) *Error[T] {
	return &Error[T]{
		logger:  logger,
		factory: factory,
		err:     err,
	}
}

func (e *Error[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

func (e Error[T]) Update(ctx context.Context, tx xio.State) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	data, pos, err := tx.Data()
	if err != nil {
		e.logger.Fatal("data error: %s", err)
	}
	if len(data) == 0 {
		err = ErrRollback
		return
	}
	level, ok := GetStateLevel(ctx)
	if !ok {
		e.logger.Fatal("state level is not set")
	}
	msg, err := e.factory.Error(ctx, level, e.err, data, int(pos), len(data))
	if err != nil {
		e.logger.Fatal("messages factory error: %s", err)
	}
	err = e.receiver.Receive(msg)
	if err != nil {
		e.logger.Fatal("receiver error: %s", err)
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Error(err error) (tail *Chain[T]) {
	defaultName := "Error"
	newNode := newError[T](b.logger, b.factory, err)
	tail = b.createNode(defaultName, func() any { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.Head().receiver)
	return
}

func isError[T any](s State[T]) (ret bool) {
	_, ret = s.(*Error[T])
	return
}
