package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type Break[T any] struct {
	logger common.Logger
}

func newBreak[T any](logger common.Logger) *Break[T] {
	return &Break[T]{
		logger: logger,
	}
}

func (b Break[T]) Update(_ context.Context, _ xio.State) (err error) {
	err = errBreak
	return
}

func (b Builder[T]) Break() (tail *Chain[T]) {
	if b.last == nil {
		b.logger.Fatal("invalid grammar: break can't be the first state in chain")
	}
	tail = b.append("Break", func() any { return newBreak[T](b.logger) })
	return
}

func isBreak[T any](state Update[T]) (ret bool) {
	_, ret = state.(*Break[T])
	return
}
