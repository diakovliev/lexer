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
	states             []State[T]
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
func (r *Run[T]) currentState() State[T] {
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

// update updates the current state of the lexer with the given transaction.
func (r *Run[T]) update(ctx context.Context, tx xio.State) (err error) {
	state := r.currentState()
	if state == nil {
		// no more states to process, we're done
		err = ErrNoMoreStates
		return
	}
	err = state.Update(ctx, tx)
	return
}

func (r *Run[T]) Run(ctx context.Context, source xio.Source) (err error) {
	// set state level
	ctx = WithNextStateLevel(ctx)
loop:
	for ctx.Err() == nil {
		tx := source.Begin()
		if err = r.update(ctx, tx); err == nil {
			r.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, ErrCommit):
			if err := tx.Commit(); err != nil {
				r.logger.Fatal("commit error: %v", err)
			}
			r.reset()
		case errors.Is(err, ErrRollback):
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %v", err)
			}
			r.next()
		case errors.Is(err, ErrNoMoreStates):
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %v", err)
			}
			if source.Has() {
				err = ErrHasMoreData
			} else {
				err = r.incompleteStateErr
			}
			break loop
		case errors.Is(err, ErrBreak):
			if err := tx.Commit(); err != nil {
				r.logger.Fatal("commit error: %v", err)
			}
			err = ErrCommit
			break loop
		default:
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %v", err)
			}
			break loop
		}
	}
	return
}
