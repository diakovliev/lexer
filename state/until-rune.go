package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// UntilRune is a state that reads until the given function returns true.
type UntilRune[T any] struct {
	logger common.Logger
	pred   RunePredicate
}

// newUntilRune creates a new state that reads until the given function returns true.
func newUntilRune[T any](logger common.Logger, pred RunePredicate) *UntilRune[T] {
	return &UntilRune[T]{
		logger: logger,
		pred:   pred,
	}
}

// Update implements the State interface. It reads until the given function returns true.
func (ur UntilRune[T]) Update(ctx context.Context, tx xio.State) (err error) {
	count := 0
	for {
		r, rw, nextErr := tx.NextRune()
		if nextErr != nil && !errors.Is(nextErr, io.EOF) {
			err = nextErr
			return
		}
		if errors.Is(nextErr, io.EOF) && rw == 0 {
			break
		}
		if ur.pred(r) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				ur.logger.Fatal("unread error: %s", unreadErr)
			}
			break
		}
		count++
	}
	if count == 0 {
		// if no runes were read, then rollback the state.
		err = errRollback
	} else {
		err = errChainNext
	}
	return
}

// UntilRune creates a state that reads runes until the pred returns true.
func (b Builder[T]) UntilRune(pred RunePredicate) (tail *Chain[T]) {
	tail = b.append("UntilRune", func() any { return newUntilRune[T](b.logger, pred) })
	return
}

// WhileRune creates a state that reads runes while the pred returns true.
func (b Builder[T]) WhileRune(pred RunePredicate) (tail *Chain[T]) {
	tail = b.append("WhileRune", func() any { return newUntilRune[T](b.logger, negatePredicate(pred)) })
	return
}
