package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Break is a state that breaks the loop and returns an error to the caller.
type Break[T any] struct {
	logger common.Logger
	action error
}

// newBreak creates a new instance of the Break state.
func newBreak[T any](logger common.Logger, action error) *Break[T] {
	return &Break[T]{
		logger: logger,
		action: action,
	}
}

// Update implements the Update interface. It advances the ioState and returns an errStateBreak error.
func (br Break[T]) Update(_ context.Context, ioState xio.State) (err error) {
	switch {
	case errors.Is(br.action, ErrCommit):
		_, _, err = ioState.Data()
		common.AssertNoError(err, "data error")
	case errors.Is(br.action, ErrRollback):
	default:
		common.AssertTrue(false, fmt.Sprintf("unknown action: %s", br.action))
	}
	err = makeErrBreak(br.action)
	return
}

// Break adds a break state to the chain.
func (b Builder[T]) Break(actions ...error) (tail *Chain[T]) {
	action := ErrCommit
	if len(actions) == 1 {
		action = actions[0]
		common.AssertErrorIsAnyFrom(action, []error{ErrCommit, ErrRollback}, fmt.Sprintf("invalid grammar: unsupported break action: %s", action))
	} else {
		common.AssertFalse(len(actions) > 1, "invalid grammar: too many actions for break")
	}
	tail = b.append("Break", func() Update[T] { return newBreak[T](b.logger, action) })
	return
}

// isBreak checks if the given state is a break state.
func isBreak[T any](state Update[T]) (ret bool) {
	_, ret = state.(*Break[T])
	return
}
