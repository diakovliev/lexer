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
	mode   fnMode
}

// NewFnRune creates a new state that checks if the next rune matches the predicate.
func newFnRune[T any](logger common.Logger, pred RunePredicate, mode fnMode) *FnRune[T] {
	return &FnRune[T]{
		logger: logger,
		pred:   pred,
		mode:   mode,
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
		err = ErrRollback
		return
	}
	result := fr.pred(r)
	switch fr.mode {
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

// isNotRepeatableFnRune returns true if the state is not repeatable
func isNotRepeatableFnRune[T any](s Update[T]) bool {
	i, ok := s.(*FnRune[T])
	if !ok {
		return false
	}
	ok = i.mode == fnLook
	return ok
}

// CheckRune is a state that matches rune by the given function.
func (b Builder[T]) CheckRune(pred RunePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("CheckRune", func() Update[T] { return newFnRune[T](b.logger, pred, fnAccept) })
	return
}

// FollowedByCheckRune is a state that matches rune by the given function and then rollbacks if it fails.
func (b Builder[T]) FollowedByCheckRune(pred RunePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("FollowedByCheckRune", func() Update[T] { return newFnRune[T](b.logger, pred, fnLook) })
	return
}

// CheckNotRune is a state that matches rune by the given function and returns an error if it does match.
func (b Builder[T]) CheckNotRune(pred RunePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("CheckNotRune", func() Update[T] { return newFnRune[T](b.logger, Not(pred), fnAccept) })
	return
}

// FollowedByCheckNotRune is a state that matches rune by the given function and rollbacks if it does match.
func (b Builder[T]) FollowedByCheckNotRune(pred RunePredicate) (tail *Chain[T]) {
	common.AssertNotNil(pred, "invalid grammar: nil predicate")
	tail = b.append("FollowedByCheckNotRune", func() Update[T] { return newFnRune[T](b.logger, Not(pred), fnLook) })
	return
}

// Rune is a state that matches the given rune.
func (b Builder[T]) Rune(sample rune) (tail *Chain[T]) {
	tail = b.append("Rune", func() Update[T] { return newFnRune[T](b.logger, IsRune(sample), fnAccept) })
	return
}

// FollowedByRune is a state that matches the given rune and rollbacks if it does not match.
func (b Builder[T]) FollowedByRune(sample rune) (tail *Chain[T]) {
	tail = b.append("FollowedByRune", func() Update[T] { return newFnRune[T](b.logger, IsRune(sample), fnLook) })
	return
}

// NotRune is a state that matches all runes except the given one.
func (b Builder[T]) NotRune(sample rune) (tail *Chain[T]) {
	tail = b.append("NotRune", func() Update[T] { return newFnRune[T](b.logger, Not(IsRune(sample)), fnAccept) })
	return
}

// FollowedByNotRune is a state that matches all runes except the given one and rollbacks if it does match.
func (b Builder[T]) FollowedByNotRune(sample rune) (tail *Chain[T]) {
	tail = b.append("FollowedByNotRune", func() Update[T] { return newFnRune[T](b.logger, Not(IsRune(sample)), fnLook) })
	return
}

// AnyRune is a state that matches any rune.
func (b Builder[T]) AnyRune() (tail *Chain[T]) {
	tail = b.append("AnyRune", func() Update[T] { return newFnRune[T](b.logger, True[rune](), fnAccept) })
	return
}

// FollowedByAnyRune is a state that matches any rune and rollbacks if it does not match.
func (b Builder[T]) FollowedByAnyRune() (tail *Chain[T]) {
	tail = b.append("FollowedByAnyRune", func() Update[T] { return newFnRune[T](b.logger, True[rune](), fnLook) })
	return
}
