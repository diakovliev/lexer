package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Provider[T any] func(b Builder[T]) []State[T]

	SubState[T any] struct {
		logger   common.Logger
		builder  Builder[T]
		provider Provider[T]
	}
)

func newSubState[T any](logger common.Logger, builder Builder[T], provider Provider[T]) *SubState[T] {
	return &SubState[T]{
		logger:   logger,
		builder:  builder,
		provider: provider,
	}
}

// Update implements State interface. It updates the current state of the lexer with the given transaction.
func (ss SubState[T]) Update(tx xio.State) (err error) {
	err = NewRun(ss.logger, ss.builder, ss.provider, ErrIncompleteState).Run(xio.AsSource(tx))
	return
}

func (b Builder[T]) State(builder Builder[T], provider Provider[T]) (head *Chain[T]) {
	defaultName := "SubState"
	head = b.createNode(defaultName, func() any { return newSubState(b.logger, builder, provider) })
	return
}
