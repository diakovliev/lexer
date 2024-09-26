package state

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidInput indicates that the input is invalid
	ErrInvalidInput = errors.New("invalid input")
	// ErrIncomplete indicates that the combined state is incomplete
	ErrIncomplete = errors.New("incomplete")

	// ErrCommit indicates that the state should be committed
	ErrCommit = errors.New("commit")
	// ErrRollback indicates that the state should be rolled back
	ErrRollback = errors.New("rollback")

	// ErrChainNext indicates that the next state inside the chain should be processed
	ErrChainNext = errors.New("next")
	// ErrChainRepeat indicates that the previous state should be repeated
	ErrChainRepeat = errors.New("repeat")

	// errStateBreak indicates that the combined state is done
	errStateBreak = errors.New("break")
	// errStateNoMoreStates indicates that there are no more states to process in the combined state
	errStateNoMoreStates = errors.New("no more states")
)

type (
	// errRepeatImpl is an implementation of the repeat error
	errRepeatImpl struct {
		q Quantifier
	}

	// errBreakImpl is an implementation of the break error
	errBreakImpl struct {
		action error
	}
)

// makeErrRepeat returns an error that can be unwrapped
func makeErrRepeat(q Quantifier) error {
	return &errRepeatImpl{
		q: q,
	}
}

// Error implements the error interface
func (errRepeatImpl) Error() string {
	return ErrChainRepeat.Error()
}

// Unwrap implements the error interface
func (err errRepeatImpl) Unwrap() error {
	return ErrChainRepeat
}

// getRepeatQuantifier returns the quantifier from an error if it is a *errRepeat
func getRepeatQuantifier(err error) (q Quantifier, ret bool) {
	if err == nil {
		return
	}
	e, ok := err.(*errRepeatImpl)
	if !ok {
		return
	}
	q = e.q
	ret = true
	return
}

func MakeErrBreak(action error) error {
	return &errBreakImpl{
		action: action,
	}
}

// Error implements the error interface
func (eb errBreakImpl) Error() string {
	return fmt.Sprintf("%s: %s", errStateBreak.Error(), eb.action)
}

// Unwrap implements the error interface
func (eb errBreakImpl) Unwrap() error {
	return errStateBreak
}

// getBreakAction returns the action from an error if it is a *errBreak
func getBreakAction(err error) (action error, ret bool) {
	if err == nil {
		return
	}
	e, ok := err.(*errBreakImpl)
	if !ok {
		return
	}
	action = e.action
	ret = true
	return
}
