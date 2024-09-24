package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Run implements base state machine for lexer.
type Run[T any] struct {
	logger   common.Logger
	builder  Builder[T]
	provider Provider[T]
	eofErr   error
	states   []Update[T]
	current  int
}

// NewRun creates a new instance of the Run state machine.
func NewRun[T any](
	logger common.Logger,
	builder Builder[T],
	provider Provider[T],
	eofErr error,
) *Run[T] {
	return &Run[T]{
		logger:   logger,
		builder:  builder,
		provider: provider,
		eofErr:   eofErr,
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

// next moves the lexer state machine to the next state.
func (r *Run[T]) next() {
	r.current++
}

// reset resets the lexer state machine to its first state.
func (r *Run[T]) reset() {
	r.current = 0
}

// update updates the current state of the lexer with the given transaction.
// It returns the io transaction associated with the state io or and lifecycle error.
func (r *Run[T]) update(ctx context.Context, source xio.Source) (tx xio.Tx, err error) {
	state := r.currentState()
	if state == nil {
		// no more states to process, we're done
		err = errStateNoMoreStates
		return
	}
	ioState := source.Begin().Ref
	tx = xio.AsTx(ioState)
	err = state.Update(ctx, ioState)
	return
}

// Run runs the lexer state machine on the given source.
func (r *Run[T]) Run(ctx context.Context, source xio.Source) (err error) {
	// set state level
	ctx = WithNextTokenLevel(ctx)
loop:
	for ctx.Err() == nil {
		var tx xio.Tx
		tx, err = r.update(ctx, source)
		if err == nil {
			r.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, errStateNoMoreStates):
			if tx != nil {
				r.logger.Fatal("unexpected not nil")
			}
			if source.Has() {
				// We're done, and we have more data to process, pass error to the parent state.
				err = ErrIncomplete
			} else {
				// We're done, and we have no more data to process.
				err = r.eofErr
			}
			break loop
		case errors.Is(err, errChainRepeat), errors.Is(err, errChainNext):
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			r.logger.Fatal("invalid grammar: repeat and next allowed only inside chain")
		case errors.Is(err, errCommit):
			if err := tx.Commit(); err != nil {
				r.logger.Fatal("commit error: %s", err)
			}
			r.reset()
		case errors.Is(err, errRollback):
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			r.next()
		case errors.Is(err, errStateBreak):
			if err := tx.Commit(); err != nil {
				r.logger.Fatal("commit error: %s", err)
			}
			err = errCommit
			break loop
		default:
			if err := tx.Rollback(); err != nil {
				r.logger.Fatal("rollback error: %s", err)
			}
			break loop
		}
	}
	return
}
