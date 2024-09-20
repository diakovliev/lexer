package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// While is a state that reads runes while fn returns true.
type While[T any] struct {
	logger common.Logger
	fn     func(rune) bool
}

func newWhile[T any](logger common.Logger, fn func(rune) bool) *While[T] {
	return &While[T]{
		logger: logger,
		fn:     fn,
	}
}

// Update implements State interface. It reads runes while fn returns true and updates the state.
func (w While[T]) Update(ctx context.Context, tx xio.State) (err error) {
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
		if !w.fn(r) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				w.logger.Fatal("unread error: %s", unreadErr)
			}
			break
		}
		count++
	}
	if count == 0 {
		// if no runes were read, then rollback the state.
		err = ErrRollback
	} else {
		err = ErrNext
	}
	return
}

// While creates a state that reads runes while fn returns true.
func (b Builder[T]) While(fn func(rune) bool) (head *Chain[T]) {
	defaultName := "While"
	head = b.createNode(defaultName, func() any { return newWhile[T](b.logger, fn) })
	return
}
