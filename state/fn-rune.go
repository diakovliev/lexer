package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// FnRune is a state that checks if the next rune matches the predicate.
type FnRune[T any] struct {
	logger common.Logger
	pred   RunePredicate
}

// NewFnRune creates a new state that checks if the next rune matches the predicate.
func newFnRune[T any](logger common.Logger, pred RunePredicate) *FnRune[T] {
	return &FnRune[T]{
		logger: logger,
		pred:   pred,
	}
}

// Update implements the Update interface. It checks if the next rune matches
// the predicate and returns an error if it doesn't match.
func (fr FnRune[T]) Update(ctx context.Context, tx xio.State) (err error) {
	r, rw, err := tx.NextRune()
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, io.EOF) && rw == 0 {
		err = errRollback
		return
	}
	if !fr.pred(r) {
		if _, unreadErr := tx.Unread(); unreadErr != nil {
			fr.logger.Fatal("unread error: %s", unreadErr)
		}
		err = errRollback
		return
	}
	err = errChainNext
	return
}

// CheckRune is a state that matches rune by the given function.
func (b Builder[T]) CheckRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FnRune", func() any { return newFnRune[T](b.logger, pred) })
	return
}

// CheckNotRune is a state that matches rune by the given function and returns an error if it does match.
func (b Builder[T]) CheckNotRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("NotFnRune", func() any { return newFnRune[T](b.logger, negatePredicate(pred)) })
	return
}

// Rune is a state that matches the given rune.
func (b Builder[T]) Rune(sample rune) (tail *Chain[T]) {
	tail = b.append("Rune", func() any { return newFnRune[T](b.logger, runeEqual(sample)) })
	return
}

// NotRune is a state that matches all runes except the given one.
func (b Builder[T]) NotRune(sample rune) (tail *Chain[T]) {
	tail = b.append("NotRune", func() any { return newFnRune[T](b.logger, negatePredicate(runeEqual(sample))) })
	return
}

// AnyRune is a state that matches any rune.
func (b Builder[T]) AnyRune() (tail *Chain[T]) {
	tail = b.append("AnyRune", func() any { return newFnRune[T](b.logger, alwaysTrue[rune]()) })
	return
}
