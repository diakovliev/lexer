package state

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	errCommit       = errors.New("commit")
	errRollback     = errors.New("rollback")
	errNext         = errors.New("next")
	errBreak        = errors.New("break")
	errRepeat       = errors.New("repeat")
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

func (errRepeatImpl) Error() string {
	return "repeat"
}

func (err errRepeatImpl) Unwrap() error {
	return errRepeat
}
