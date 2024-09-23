package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Provider[T any] func(b Builder[T]) []Update[T]
	StateFinalize   func(ctx context.Context, logger common.Logger, result error) error

	State[T any] struct {
		logger   common.Logger
		builder  Builder[T]
		provider Provider[T]
		finalize StateFinalize
	}
)

func newState[T any](logger common.Logger, builder Builder[T], provider Provider[T], finalize StateFinalize) *State[T] {
	return &State[T]{
		logger:   logger,
		builder:  builder,
		provider: provider,
		finalize: finalize,
	}
}

// Update implements State interface. It updates the current state of the lexer with the given transaction.
func (s State[T]) Update(ctx context.Context, tx xio.State) (err error) {
	err = NewRun(s.logger, s.builder, s.provider, ErrIncompleteState).Run(ctx, xio.AsSource(tx))
	err = s.finalize(ctx, s.logger, err)
	return
}

func errorPassthrough(_ context.Context, _ common.Logger, result error) (err error) {
	err = result
	return
}

func errorInverse(ctx context.Context, logger common.Logger, result error) (err error) {
	err = result
	if !errors.Is(err, ErrCommit) && errors.Is(err, ErrRollback) {
		return
	}
	switch {
	case errors.Is(err, ErrCommit):
		err = ErrRollback
	case errors.Is(err, ErrRollback):
		err = ErrCommit
	default:
		logger.Fatal("unreachable code")
	}
	return
}

// State creates a new state that will be used to update the current state of the lexer.
// It returns the tail of the chain.
func (b Builder[T]) State(builder Builder[T], provider Provider[T]) (tail *Chain[T]) {
	defaultName := "State"
	tail = b.createNode(defaultName, func() any { return newState(b.logger, builder, provider, errorPassthrough) })
	return
}

// NotState creates a new state that will be used to update the current state of the lexer.
// It returns the tail of the chain.
func (b Builder[T]) NotState(builder Builder[T], provider Provider[T]) (tail *Chain[T]) {
	defaultName := "NotState"
	tail = b.createNode(defaultName, func() any { return newState(b.logger, builder, provider, errorInverse) })
	return
}
