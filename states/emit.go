package states

import (
	"github.com/diakovliev/lexer/common"
)

type Emit[T any] struct {
	logger      common.Logger
	messageType common.MessageType
	userType    T
	receiver    common.Receiver[T]
}

func newEmit[T any](
	logger common.Logger,
	messageType common.MessageType,
	userType T,
) *Emit[T] {
	return &Emit[T]{
		logger:      logger,
		messageType: messageType,
		userType:    userType,
	}
}

func (e *Emit[T]) SetReceiver(receiver common.Receiver[T]) {
	e.receiver = receiver
}

func (e Emit[T]) Update(tx common.ReadUnreadData) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	pos, data, err := tx.Data()
	if err != nil {
		e.logger.Error("tx.Data() = data=%v, err=%s", data, err)
		return
	}
	if len(data) == 0 {
		e.logger.Fatal("nothing to emit")
		return
	}
	err = e.receiver(common.Message[T]{
		Type:     e.messageType,
		UserType: e.userType,
		Value:    data,
		Pos:      int(pos),
		Width:    len(data),
	})
	if err != nil {
		e.logger.Error("e.receiver() = err=%s", err)
		return
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Emit(messageType common.MessageType, userType T) (head *Chain[T]) {
	defaultName := "Emit"
	newNode := newEmit(b.logger, messageType, userType)
	head = b.createNode(defaultName, func() any { return newNode })
	// sent all messages to the receiver of the first node
	newNode.SetReceiver(head.receiver)
	return
}
