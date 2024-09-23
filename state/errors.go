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
	// errNext indicates that the next state inside the chain should be processed
	errNext = errors.New("next")
	// errBreak indicates that the combined state is done
	errBreak = errors.New("break")
	// errRepeat indicates that the previous state should be repeated
	errRepeat = errors.New("repeat")
	// errNoMoreStates indicates that there are no more states to process in the combined state
	errNoMoreStates = errors.New("no more states")
)

type errRepeatImpl struct {
	Q Quantifier
}

// makeErrRepeat returns an error that can be unwrapped
func makeErrRepeat(Q Quantifier) error {
	return &errRepeatImpl{
		Q: Q,
	}
}

// getQuantifier returns the quantifier from an error if it is a *errRepeat
func getQuantifier(err error) (Quantifier, bool) {
	if err == nil {
		return Quantifier{}, false
	}
	e, ok := err.(*errRepeatImpl)
	if !ok {
		return Quantifier{}, false
	}
	return e.Q, true
}

// Error implements the error interface
func (errRepeatImpl) Error() string {
	return "repeat"
}

// Unwrap implements the error interface
func (err errRepeatImpl) Unwrap() error {
	return errRepeat
}
