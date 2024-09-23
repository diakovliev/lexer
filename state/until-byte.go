package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// UntilRune is a state that reads until the given function returns true.
type UntilByte[T any] struct {
	logger common.Logger
	pred   BytePredicate
}

func newUntilByte[T any](logger common.Logger, pred BytePredicate) *UntilByte[T] {
	return &UntilByte[T]{
		logger: logger,
		pred:   pred,
	}
}

// Update implements the State interface. It reads until the given function returns true.
func (ub UntilByte[T]) Update(ctx context.Context, tx xio.State) (err error) {
	count := 0
	for {
		b, nextErr := tx.NextByte()
		if nextErr != nil && !errors.Is(nextErr, io.EOF) {
			err = nextErr
			return
		}
		if errors.Is(nextErr, io.EOF) {
			break
		}
		if ub.pred(b) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				ub.logger.Fatal("unread error: %s", unreadErr)
			}
			break
		}
		count++
	}
	if count == 0 {
		// if no runes were read, then rollback the state.
		err = errRollback
	} else {
		err = errNext
	}
	return
}

// UntilRune creates a state that reads bytes until the pred returns true.
func (b Builder[T]) UntilByte(pred BytePredicate) (tail *Chain[T]) {
	defaultName := "UntilByte"
	tail = b.createNode(defaultName, func() any { return newUntilByte[T](b.logger, pred) })
	return
}

// WhileByte creates a state that reads bytes while the pred returns true.
func (b Builder[T]) WhileByte(pred BytePredicate) (tail *Chain[T]) {
	defaultName := "WhileByte"
	tail = b.createNode(defaultName, func() any { return newUntilByte[T](b.logger, negatePredicate(pred)) })
	return
}
