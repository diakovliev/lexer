package states

import (
	"github.com/diakovliev/lexer/common"
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
		receiver common.Receiver[T]
	}
)

func newError[T any](logger common.Logger, err error) *Error[T] {
	return &Error[T]{
		logger: logger,
		err:    err,
	}
}

func (e *Error[T]) SetReceiver(receiver common.Receiver[T]) {
	e.receiver = receiver
}

func (e Error[T]) Update(tx xio.State) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	data, pos, err := tx.Data()
	if err != nil {
		e.logger.Error("tx.Data() = data=%v, err=%s", data, err)
		return
	}
	if len(data) == 0 {
		err = ErrRollback
		return
	}
	err = e.receiver(common.Message[T]{
		Type: common.Error,
		Value: &ErrorValue{
			Err:   e.err,
			Value: data,
		},
		Pos:   int(pos),
		Width: len(data),
	})
	if err != nil {
		e.logger.Error("e.receiver() = err=%s", err)
		return
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Error(err error) (head *Chain[T]) {
	defaultName := "Error"
	newNode := newError[T](b.logger, err)
	head = b.createNode(defaultName, func() any { return newNode })
	// sent all messages to the receiver of the first node
	newNode.SetReceiver(head.receiver)
	return
}
