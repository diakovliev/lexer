package states

import "github.com/diakovliev/lexer/common"

type Break[T any] struct {
	logger common.Logger
}

func newBreak[T any](logger common.Logger) *Break[T] {
	return &Break[T]{
		logger: logger,
	}
}

func (b Break[T]) Update(_ common.ReadUnreadData) (err error) {
	err = errBreak
	// b.logger.Fatal("BREAK")
	return
}

func (b Builder[T]) Break() (head *Chain[T]) {
	defaultName := "Break"
	head = b.createNode(defaultName, func() any { return newBreak[T](b.logger) })
	return
}
