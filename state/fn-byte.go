package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// FnRune is a state that checks if the next rune matches the predicate.
type FnByte[T any] struct {
	logger common.Logger
	pred   BytePredicate
	mode   fnMode
}

// newFnRune creates a new state that checks if the next rune matches the predicate.
func newFnByte[T any](logger common.Logger, pred BytePredicate, mode fnMode) *FnByte[T] {
	return &FnByte[T]{
		logger: logger,
		pred:   pred,
		mode:   mode,
	}
}

// Update implements the Update interface. It checks if the next rune matches
// the predicate and returns an error if it doesn't match.
func (fb FnByte[T]) Update(ctx context.Context, tx xio.State) (err error) {
	b, err := tx.NextByte()
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, io.EOF) {
		err = ErrRollback
		return
	}
	result := fb.pred(b)
	switch fb.mode {
	case fnAccept:
		if !result {
			_, err = tx.Unread()
			common.AssertNoError(err, "unread error")
			err = ErrRollback
			return
		}
	case fnLook:
		_, err = tx.Unread()
		common.AssertNoError(err, "unread error")
	}
	if result {
		err = ErrChainNext
	} else {
		err = ErrRollback
	}
	return
}

// isNotRepeatableFnByte returns true if the state is not repeatable
func isNotRepeatableFnByte[T any](s Update[T]) bool {
	i, ok := s.(*FnByte[T])
	if !ok {
		return false
	}
	ok = i.mode == fnLook
	return ok
}

// ByteCheck adds a state that checks if the next rune matches the predicate to the chain.
func (b Builder[T]) ByteCheck(pred BytePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("ByteCheck", func() Update[T] { return newFnByte[T](b.logger, pred, fnAccept) })
	return
}

// FollowedByByteCheck adds a state that checks if the next rune matches the predicate to the chain, and rolls back if it does not match.
func (b Builder[T]) FollowedByByteCheck(pred BytePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("FollowedByByteCheck", func() Update[T] { return newFnByte[T](b.logger, pred, fnLook) })
	return
}

// NotByteCheck adds a state that checks if the next rune doesn't match the predicate to the chain.
func (b Builder[T]) NotByteCheck(pred BytePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("NotByteCheck", func() Update[T] { return newFnByte[T](b.logger, Not(pred), fnAccept) })
	return
}

// FollowedByNotByteCheck adds a state that checks if the next rune doesn't match the predicate to the chain.
func (b Builder[T]) FollowedByNotByteCheck(pred BytePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("FollowedByNotByteCheck", func() Update[T] { return newFnByte[T](b.logger, Not(pred), fnLook) })
	return
}

// Byte adds a state that checks if the next rune matches the sample to the chain.
func (b Builder[T]) Byte(sample byte) (tail *Chain[T]) {
	tail = b.append("Byte", func() Update[T] { return newFnByte[T](b.logger, IsByte(sample), fnAccept) })
	return
}

// FollowedByByte adds a state that checks if the next rune matches the sample to the chain.
func (b Builder[T]) FollowedByByte(sample byte) (tail *Chain[T]) {
	tail = b.append("FollowedByByte", func() Update[T] { return newFnByte[T](b.logger, IsByte(sample), fnLook) })
	return
}

// NotByte adds a state that checks if the next rune doesn't match the sample to the chain.
func (b Builder[T]) NotByte(sample byte) (tail *Chain[T]) {
	tail = b.append("NotByte", func() Update[T] { return newFnByte[T](b.logger, Not(IsByte(sample)), fnAccept) })
	return
}

// FollowedByNotByte adds a state that checks if the next rune doesn't match the sample to the chain.
func (b Builder[T]) FollowedByNotByte(sample byte) (tail *Chain[T]) {
	tail = b.append("FollowedByNotByte", func() Update[T] { return newFnByte[T](b.logger, Not(IsByte(sample)), fnLook) })
	return
}

// AnyByte adds a state that accepts any byte to the chain.
func (b Builder[T]) AnyByte() (tail *Chain[T]) {
	tail = b.append("AnyByte", func() Update[T] { return newFnByte[T](b.logger, True[byte](), fnAccept) })
	return
}

// FollowedByAnyByte adds a state that checks if the next rune matches any byte to the chain.
func (b Builder[T]) FollowedByAnyByte() (tail *Chain[T]) {
	tail = b.append("FollowedByAnyByte", func() Update[T] { return newFnByte[T](b.logger, True[byte](), fnLook) })
	return
}
