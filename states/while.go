package states

import (
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
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
func (w While[T]) Update(tx common.ReadUnreadData) (err error) {
	count := 0
	for {
		data, r, nextErr := common.NextRuneFrom(tx)
		if nextErr != nil && !errors.Is(nextErr, io.EOF) {
			err = nextErr
			return
		}
		if errors.Is(nextErr, io.EOF) && len(data) == 0 {
			break
		}
		if !w.fn(r) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				w.logger.Error("unread error: %s", unreadErr)
			}
			break
		}
		count++
	}
	if count == 0 {
		// if no runes were read, then rollback the state.
		err = ErrRollback
	} else {
		err = errChainNext
	}
	return
}

// While creates a state that reads runes while fn returns true.
func (b Builder[T]) While(fn func(rune) bool) (head *Chain[T]) {
	defaultName := "While"
	head = b.createNode(defaultName, func() any { return newWhile[T](b.logger, fn) })
	return
}
