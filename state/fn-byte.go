package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	FnByte[T any] struct {
		logger common.Logger
		pred   BytePredicate
	}
)

func newFnByte[T any](logger common.Logger, pred BytePredicate) *FnByte[T] {
	return &FnByte[T]{
		logger: logger,
		pred:   pred,
	}
}

func (fb FnByte[T]) Update(ctx context.Context, tx xio.State) (err error) {
	b, err := tx.NextByte()
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if (errors.Is(err, io.EOF)) || !fb.pred(b) {
		err = ErrRollback
		return
	}
	err = ErrNext
	return
}

func (b Builder[T]) FnByte(pred BytePredicate) (tail *Chain[T]) {
	tail = b.createNode("FnByte", func() any { return newFnByte[T](b.logger, pred) })
	return
}

func (b Builder[T]) NotFbByte(pred BytePredicate) (tail *Chain[T]) {
	tail = b.createNode("NotFnByte", func() any { return newFnByte[T](b.logger, negatePredicate(pred)) })
	return
}

func (b Builder[T]) Byte(sample byte) (tail *Chain[T]) {
	tail = b.createNode("Byte", func() any { return newFnByte[T](b.logger, byteEqual(sample)) })
	return
}

func (b Builder[T]) NotByte(sample byte) (tail *Chain[T]) {
	tail = b.createNode("NotByte", func() any { return newFnByte[T](b.logger, negatePredicate(byteEqual(sample))) })
	return
}

func (b Builder[T]) AnyByte() (tail *Chain[T]) {
	tail = b.createNode("AnyByte", func() any { return newFnByte[T](b.logger, alwaysTrue[byte]()) })
	return
}
