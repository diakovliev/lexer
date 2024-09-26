package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Provider is a function that returns collection of states
	Provider[T any] func(b Builder[T]) []Update[T]

	// State is a combined state
	State[T any] struct {
		run *Run[T]
	}
)

// newState creates a new instance of State
func newState[T any](logger common.Logger, builder Builder[T], provider Provider[T]) *State[T] {
	return &State[T]{
		run: NewRun(logger, builder, provider, ErrInvalidInput),
	}
}

// Update implements State interface. It updates the current state of the lexer with the given transaction.
func (s *State[T]) Update(ctx context.Context, tx xio.State) (err error) {
	err = s.run.Run(ctx, xio.AsSource(tx))
	defer s.run.Reset()
	return
}

// State creates a new state that will be used to update the current state of the lexer.
// It returns the tail of the chain.
func (b Builder[T]) State(builder Builder[T], provider Provider[T]) (tail *Chain[T]) {
	tail = b.append("State", func() Update[T] { return newState(b.logger, builder, provider) })
	return
}
