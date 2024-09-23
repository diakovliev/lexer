package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Provider[T any] func(b Builder[T]) []Update[T]

	State[T any] struct {
		logger   common.Logger
		builder  Builder[T]
		provider Provider[T]
	}
)

func newState[T any](logger common.Logger, builder Builder[T], provider Provider[T]) *State[T] {
	return &State[T]{
		logger:   logger,
		builder:  builder,
		provider: provider,
	}
}

// Update implements State interface. It updates the current state of the lexer with the given transaction.
func (s State[T]) Update(ctx context.Context, tx xio.State) (err error) {
	err = NewRun(s.logger, s.builder, s.provider, ErrIncompleteState).Run(ctx, xio.AsSource(tx))
	return
}

// State creates a new state that will be used to update the current state of the lexer.
// It returns the tail of the chain.
func (b Builder[T]) State(builder Builder[T], provider Provider[T]) (tail *Chain[T]) {
	tail = b.createNode("State", func() any { return newState(b.logger, builder, provider) })
	return
}
