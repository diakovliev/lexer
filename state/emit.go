package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type Emit[T any] struct {
	logger   common.Logger
	token    T
	factory  message.Factory[T]
	receiver message.Receiver[T]
}

func newEmit[T any](
	logger common.Logger,
	factory message.Factory[T],
	token T,
) *Emit[T] {
	return &Emit[T]{
		logger:  logger,
		factory: factory,
		token:   token,
	}
}

func (e *Emit[T]) setReceiver(receiver message.Receiver[T]) {
	e.receiver = receiver
}

func (e Emit[T]) Update(ctx context.Context, tx xio.State) (err error) {
	if e.receiver == nil {
		e.logger.Fatal("receiver is not set")
		return
	}
	data, pos, err := tx.Data()
	if err != nil {
		e.logger.Fatal("data error: %s", err)
	}
	if len(data) == 0 {
		e.logger.Fatal("nothing to emit")
	}
	level, ok := GetTokenLevel(ctx)
	if !ok {
		e.logger.Fatal("no token level in context")
	}
	msg, err := e.factory.Token(ctx, level, e.token, data, int(pos), len(data))
	if err != nil {
		e.logger.Fatal("messages factory error: %s", err)
	}
	err = e.receiver.Receive(msg)
	if err != nil {
		e.logger.Fatal("receiver error: %s", err)
	}
	err = errCommit
	return
}

func (b Builder[T]) Emit(token T) (tail *Chain[T]) {
	if b.last == nil {
		b.logger.Fatal("invalid grammar: emit can't be the first state in chain")
	}
	newNode := newEmit(b.logger, b.factory, token)
	tail = b.append("Emit", func() any { return newNode })
	// sent all messages to the the first node receiver
	newNode.setReceiver(tail.Head().receiver)
	return
}

func isEmit[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Emit[T])
	return
}
