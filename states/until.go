package states

import (
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
)

// Until is a state that reads until the given function returns true.
type Until[T any] struct {
	logger common.Logger
	fn     func(rune) bool
}

func newUntil[T any](logger common.Logger, fn func(rune) bool) *Until[T] {
	return &Until[T]{
		logger: logger,
		fn:     fn,
	}
}

// Update implements the State interface. It reads until the given function returns true.
func (u Until[T]) Update(tx common.ReadUnreadData) (err error) {
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
		if u.fn(r) {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				u.logger.Error("unread error: %s", unreadErr)
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

// Until creates a state that reads until the given function returns true.
func (b Builder[T]) Until(fn func(rune) bool) (head *Chain[T]) {
	defaultName := "Until"
	head = b.createNode(defaultName, func() any { return newUntil[T](b.logger, fn) })
	return
}
