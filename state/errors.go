package state

import "errors"

var (
	ErrNoMoreStates    = errors.New("no more states")
	ErrIncompleteState = errors.New("incomplete state")
	ErrHasMoreData     = errors.New("has more data")
	ErrCommit          = errors.New("commit")
	ErrRollback        = errors.New("rollback")
	ErrNext            = errors.New("next")
	ErrBreak           = errors.New("break")
	ErrRepeat          = errors.New("repeat")
)

type errRepeat struct {
	Q Quantifier
}

// MakeRepeat returns an error that can be unwrapped
func MakeRepeat(Q Quantifier) error {
	return &errRepeat{
		Q: Q,
	}
}

// getQuantifier returns the quantifier from an error if it is a *errRepeat
func getQuantifier(err error) (Quantifier, bool) {
	if err == nil {
		return Quantifier{}, false
	}
	e, ok := err.(*errRepeat)
	if !ok {
		return Quantifier{}, false
	}
	return e.Q, true
}

func (errRepeat) Error() string {
	return "repeat"
}

func (err errRepeat) Unwrap() error {
	return ErrRepeat
}
