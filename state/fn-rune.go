package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	FnRune[T any] struct {
		logger common.Logger
		pred   RunePredicate
	}
)

func newFnRune[T any](logger common.Logger, pred RunePredicate) *FnRune[T] {
	return &FnRune[T]{
		logger: logger,
		pred:   pred,
	}
}

func (fr FnRune[T]) Update(ctx context.Context, tx xio.State) (err error) {
	r, rw, err := tx.NextRune()
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if (errors.Is(err, io.EOF) && rw == 0) || !fr.pred(r) {
		err = ErrRollback
		return
	}
	err = ErrNext
	return
}

// FnRune is a state that matches rune by the given function.
func (b Builder[T]) FnRune(pred RunePredicate) (tail *Chain[T]) {
	tail = b.createNode("FnRune", func() any { return newFnRune[T](b.logger, pred) })
	return
}

// Rune is a state that matches the given rune.
func (b Builder[T]) Rune(sample rune) (tail *Chain[T]) {
	tail = b.createNode("Rune", func() any { return newFnRune[T](b.logger, func(r rune) bool { return r == sample }) })
	return
}

// NotRune is a state that matches all runes except the given one.
func (b Builder[T]) NotRune(sample rune) (tail *Chain[T]) {
	tail = b.createNode("NotRune", func() any { return newFnRune[T](b.logger, func(r rune) bool { return r != sample }) })
	return
}

// AnyRune is a state that matches any rune.
func (b Builder[T]) AnyRune() (tail *Chain[T]) {
	tail = b.createNode("AnyRune", func() any { return newFnRune[T](b.logger, func(_ rune) bool { return true }) })
	return
}
