package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type Run[T any] struct {
	logger             common.Logger
	builder            Builder[T]
	provider           Provider[T]
	incompleteStateErr error
	states             []Update[T]
	current            int
}

func NewRun[T any](
	logger common.Logger,
	builder Builder[T],
	provider Provider[T],
	incompleteStateErr error,
) *Run[T] {
	return &Run[T]{
		logger:             logger,
		builder:            builder,
		provider:           provider,
		incompleteStateErr: incompleteStateErr,
	}
}

// currentState returns the current state of the lexer.
func (r *Run[T]) currentState() Update[T] {
	if len(r.states) == 0 && r.provider != nil {
		r.states = r.provider(r.builder)
	}
	if len(r.states) == 0 {
		return nil
	}
	if len(r.states) <= r.current {
		return nil
	}
	return r.states[r.current]
}

// next moves the lexer to the next state.
func (r *Run[T]) next() {
	r.current++
}

// reset resets the lexer to its first state.
func (r *Run[T]) reset() {
	r.current = 0
}

func (r *Run[T]) isLast() (ret bool) {
	return r.current == len(r.states)-1
}

// update updates the current state of the lexer with the given transaction.
func (r *Run[T]) update(ctx context.Context, source xio.Source) (tx xio.State, err error) {
	state := r.currentState()
	if state == nil {
		// no more states to process, we're done
		err = ErrNoMoreStates
		return
	}
	tx = source.Begin().Ref
	err = state.Update(ctx, tx)
	return
}

func (r *Run[T]) Run(ctx context.Context, source xio.Source) (err error) {
	// set state level
	ctx = WithNextTokenLevel(ctx)
loop:
	for ctx.Err() == nil {
		var tx xio.State
		tx, err = r.update(ctx, source)
		if err == nil {
			r.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, ErrNoMoreStates):
			if !source.Has() {
				err = r.incompleteStateErr
			}
			break loop
		case errors.Is(err, ErrRepeat), errors.Is(err, ErrNext):
			if err := xio.AsTx(tx).Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			r.logger.Fatal("invalid grammar: repeat and next allowed only inside chain")
		case errors.Is(err, ErrCommit):
			if err := xio.AsTx(tx).Commit(); err != nil {
				r.logger.Fatal("commit error: %s", err)
			}
			r.reset()
		case errors.Is(err, ErrRollback):
			if err := xio.AsTx(tx).Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			r.next()
		case errors.Is(err, ErrBreak):
			if err := xio.AsTx(tx).Commit(); err != nil {
				r.logger.Fatal("commit error: %s", err)
			}
			err = ErrCommit
			break loop
		default:
			if err := xio.AsTx(tx).Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			break loop
		}
	}
	return
}
