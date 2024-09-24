package state

import "errors"

var (
	// ErrInvalidInput indicates that the input is invalid
	ErrInvalidInput = errors.New("invalid input")
	// ErrIncomplete indicates that the combined state is incomplete
	ErrIncomplete = errors.New("incomplete")

	// errCommit indicates that the state should be committed
	errCommit = errors.New("commit")
	// errRollback indicates that the state should be rolled back
	errRollback = errors.New("rollback")

	// errChainNext indicates that the next state inside the chain should be processed
	errChainNext = errors.New("next")
	// errChainRepeat indicates that the previous state should be repeated
	errChainRepeat = errors.New("repeat")

	// errStateBreak indicates that the combined state is done
	errStateBreak = errors.New("break")
	// errStateNoMoreStates indicates that there are no more states to process in the combined state
	errStateNoMoreStates = errors.New("no more states")
)

// errRepeatImpl is an implementation of the repeat error
type errRepeatImpl struct {
	q Quantifier
}

// makeErrRepeat returns an error that can be unwrapped
func makeErrRepeat(q Quantifier) error {
	return &errRepeatImpl{
		q: q,
	}
}

// Error implements the error interface
func (errRepeatImpl) Error() string {
	return errChainRepeat.Error()
}

// Unwrap implements the error interface
func (err errRepeatImpl) Unwrap() error {
	return errChainRepeat
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
