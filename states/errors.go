package states

import "errors"

var (
	ErrNoMoreStates       = errors.New("no more states")
	ErrIncompleteSubState = errors.New("incomplete sub state")
	ErrHasMoreData        = errors.New("has more data")
	ErrCommit             = errors.New("commit")
	ErrRollback           = errors.New("rollback")
	errChainNext          = errors.New("chain next")
	errBreak              = errors.New("break")
)
