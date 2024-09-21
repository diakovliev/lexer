package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	ErrorValue struct {
		Err   error
		Value any
	}

	Error[T any] struct {
		logger   common.Logger
		err      error
		receiver message.Receiver[T]
	}
)

func newError[T any](logger common.Logger, err error) *Error[T] {
	return &Error[T]{
		logger: logger,
		err:    err,
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
	err = e.receiver.Receive(message.Message[T]{
		Level: level,
		Type:  message.Error,
		Value: &ErrorValue{
			Err:   e.err,
			Value: data,
		},
		Pos:   int(pos),
		Width: len(data),
	})
	if err != nil {
		e.logger.Fatal("receiver error: %s", err)
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Error(err error) (tail *Chain[T]) {
	defaultName := "Error"
	newNode := newError[T](b.logger, err)
	tail = b.createNode(defaultName, func() any { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.Head().receiver)
	return
}

func isError[T any](s State[T]) (ret bool) {
	_, ret = s.(*Error[T])
	return
}
