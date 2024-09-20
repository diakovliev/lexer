package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type Emit[T any] struct {
	logger      common.Logger
	messageType message.MessageType
	userType    T
	receiver    message.Receiver[T]
}

func newEmit[T any](
	logger common.Logger,
	messageType message.MessageType,
	userType T,
) *Emit[T] {
	return &Emit[T]{
		logger:      logger,
		messageType: messageType,
		userType:    userType,
	}
}

func (e *Emit[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

func (e Emit[T]) Update(tx xio.State) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	data, pos, err := tx.Data()
	if err != nil {
		e.logger.Fatal("data error: %s", err)
	}
	// FIXME: do we need this check?
	if len(data) == 0 {
		e.logger.Fatal("nothing to emit")
	}
	err = e.receiver.Receive(message.Message[T]{
		Type:     e.messageType,
		UserType: e.userType,
		Value:    data,
		Pos:      int(pos),
		Width:    len(data),
	})
	if err != nil {
		e.logger.Fatal("receiver error: %s", err)
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Emit(messageType message.MessageType, userType T) (head *Chain[T]) {
	defaultName := "Emit"
	newNode := newEmit(b.logger, messageType, userType)
	head = b.createNode(defaultName, func() any { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(head.Receiver)
	return
}
