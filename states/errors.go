package states

import "errors"

var (
	ErrNoMoreStates    = errors.New("no more states")
	ErrIncompleteState = errors.New("incomplete state")
	ErrHasMoreData     = errors.New("has more data")
	ErrCommit          = errors.New("commit")
	ErrRollback        = errors.New("rollback")
	errNext            = errors.New("next")
	errBreak           = errors.New("break")
)
