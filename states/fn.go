package states

import (
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
)

type Fn[T any] struct {
	logger common.Logger
	fn     func(rune) bool
}

func newFn[T any](logger common.Logger, fn func(rune) bool) *Fn[T] {
	return &Fn[T]{
		logger: logger,
		fn:     fn,
	}
}

func (f Fn[T]) Update(tx common.ReadUnreadData) (err error) {
	data, r, err := common.NextRuneFrom(tx)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, io.EOF) && len(data) == 0 {
		err = ErrRollback
		return
	}
	if !f.fn(r) {
		err = ErrRollback
		return
	}
	err = errChainNext
	return
}

// Fn is a state that matches rune by the given function.
func (b Builder[T]) Fn(fn func(rune) bool) (head *Chain[T]) {
	defaultName := "Fn"
	head = b.createNode(defaultName, func() any { return newFn[T](b.logger, fn) })
	return
}

// Rune is a state that matches the given rune.
func (b Builder[T]) Rune(ir rune) (head *Chain[T]) {
	defaultName := "Rune"
	head = b.createNode(defaultName, func() any { return newFn[T](b.logger, func(r rune) bool { return r == ir }) })
	return
}

// AnyRune is a state that matches any rune.
func (b Builder[T]) AnyRune() (head *Chain[T]) {
	defaultName := "AnyRune"
	head = b.createNode(defaultName, func() any { return newFn[T](b.logger, func(r rune) bool { return true }) })
	return
}
