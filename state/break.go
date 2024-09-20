package state

import (
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

func (b Break[T]) Update(_ xio.State) (err error) {
	err = ErrBreak
	return
}

func (b Builder[T]) Break() (head *Chain[T]) {
	defaultName := "Break"
	head = b.createNode(defaultName, func() any { return newBreak[T](b.logger) })
	return
}
