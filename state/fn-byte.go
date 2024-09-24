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
}

// newFnRune creates a new state that checks if the next rune matches the predicate.
func newFnByte[T any](logger common.Logger, pred BytePredicate) *FnByte[T] {
	return &FnByte[T]{
		logger: logger,
		pred:   pred,
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
		err = errRollback
		return
	}
	if !fb.pred(b) {
		if _, unreadErr := tx.Unread(); unreadErr != nil {
			fb.logger.Fatal("unread error: %s", unreadErr)
		}
		err = errRollback
		return
	}
	err = errChainNext
	return
}

// CheckByte adds a state that checks if the next rune matches the predicate to the chain.
func (b Builder[T]) CheckByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FnByte", func() any { return newFnByte[T](b.logger, pred) })
	return
}

// CheckNotByte adds a state that checks if the next rune doesn't match the predicate to the chain.
func (b Builder[T]) CheckNotByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("NotFnByte", func() any { return newFnByte[T](b.logger, negatePredicate(pred)) })
	return
}

// Byte adds a state that checks if the next rune matches the sample to the chain.
func (b Builder[T]) Byte(sample byte) (tail *Chain[T]) {
	tail = b.append("Byte", func() any { return newFnByte[T](b.logger, byteEqual(sample)) })
	return
}

// NotByte adds a state that checks if the next rune doesn't match the sample to the chain.
func (b Builder[T]) NotByte(sample byte) (tail *Chain[T]) {
	tail = b.append("NotByte", func() any { return newFnByte[T](b.logger, negatePredicate(byteEqual(sample))) })
	return
}

// ByteRange adds a state that checks if the next rune matches the range to the chain.
func (b Builder[T]) AnyByte() (tail *Chain[T]) {
	tail = b.append("AnyByte", func() any { return newFnByte[T](b.logger, alwaysTrue[byte]()) })
	return
}
