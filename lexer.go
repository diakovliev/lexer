package lexer

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/states"
)

type (
	// Lexer is a lexical analyzer that reads input data and produces tokens.
	Lexer[T any] struct {
		logger   common.Logger
		reader   *common.Reader
		states   []states.State[T]
		messages []common.Message[T]
		current  int
	}

	// StatesProvider is a function that returns slice of states.
	StatesProvider[T any] func(factory states.Builder[T]) (states []states.State[T])
)

// New creates a new lexer instance with the given reader and logger.
func New[T any](logger common.Logger, reader io.Reader) (ret *Lexer[T]) {
	ret = &Lexer[T]{
		logger:  logger,
		reader:  common.NewReader(logger, reader),
		current: 0,
	}
	return ret
}

// receiver is a function that receives messages from the lexer.
func (l *Lexer[T]) receiver(m common.Message[T]) (err error) {
	l.messages = append(l.messages, m)
	return nil
}

// Messages returns all messages produced by the lexer.
func (l Lexer[T]) Messages() []common.Message[T] {
	return l.messages
}

// With adds a new states produced by given provider to the lexer.
func (l *Lexer[T]) With(fn StatesProvider[T]) *Lexer[T] {
	l.states = append(l.states, fn(states.Make(l.logger, l.receiver))...)
	return l
}

// currentState returns the current state of the lexer.
func (l *Lexer[T]) currentState() states.State[T] {
	if len(l.states) == 0 {
		return nil
	}
	if len(l.states) <= l.current {
		return nil
	}
	return l.states[l.current]
}

// next moves the lexer to the next state.
func (l *Lexer[T]) next() {
	l.current++
}

// reset resets the lexer to its first state.
func (l *Lexer[T]) reset() {
	l.current = 0
}

// update updates the current state of the lexer with the given transaction.
func (l *Lexer[T]) update(tx common.ReadUnreadData) (err error) {
	state := l.currentState()
	if state == nil {
		// no more states to process, we're done
		err = states.ErrNoMoreStates
		return
	}
	err = state.Update(tx)
	return
}

// Run runs the lexer until it is done or an error occurs.
func (l *Lexer[T]) Run(ctx context.Context) (err error) {
	l.logger.Trace("=>> enter Run()")
	defer func() { l.logger.Trace("<<= leave Run() = err=%s", err) }()
loop:
	for ctx.Err() == nil {
		tx := l.reader.Begin()
		if err = l.update(tx); err == nil {
			l.logger.Fatal("unexpected nil")
		}
		switch {
		case errors.Is(err, states.ErrCommit):
			l.logger.Trace("ErrCommit")
			if _, commitErr := tx.Commit(); commitErr != nil {
				l.logger.Error("ErrCommit -> commit error: %v", commitErr)
				err = commitErr
				break loop
			}
			l.reset()
		case errors.Is(err, states.ErrRollback):
			l.logger.Trace("ErrRollback")
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.logger.Error("ErrRollback -> rollback error: %v", rollbackErr)
				err = rollbackErr
				break loop
			}
			l.next()
		case errors.Is(err, states.ErrNoMoreStates):
			l.logger.Trace("ErrNoMoreStates")
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.logger.Error("ErrNoMoreStates -> rollback error: %v", rollbackErr)
				err = rollbackErr
				break loop
			}
			if l.reader.Has() {
				l.logger.Error("has non processed data")
				err = states.ErrHasMoreData
			} else {
				l.logger.Trace("EOF")
				err = io.EOF
			}
			break loop
		default:
			l.logger.Error("unexpected error: %v", err)
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				l.logger.Error("%s -> rollback error: %v", err, rollbackErr)
				err = rollbackErr
			}
			break loop
		}
	}
	return
}
