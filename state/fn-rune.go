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
	mode   FnMode
}

// NewFnRune creates a new state that checks if the next rune matches the predicate.
func newFnRune[T any](logger common.Logger, pred RunePredicate, mode FnMode) *FnRune[T] {
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
	case FnAccept:
		if !result {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				fr.logger.Fatal("unread error: %s", unreadErr)
			}
			err = ErrRollback
			return
		}
	case FnLook:
		if _, unreadErr := tx.Unread(); unreadErr != nil {
			fr.logger.Fatal("unread error: %s", unreadErr)
		}
	}
	if result {
		err = errChainNext
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
	ok = i.mode == FnLook
	return ok
}

// CheckRune is a state that matches rune by the given function.
func (b Builder[T]) CheckRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FnRune", func() Update[T] { return newFnRune[T](b.logger, pred, FnAccept) })
	return
}

func (b Builder[T]) FollowedByCheckRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FollowedByCheckRune", func() Update[T] { return newFnRune[T](b.logger, pred, FnLook) })
	return
}

// CheckNotRune is a state that matches rune by the given function and returns an error if it does match.
func (b Builder[T]) CheckNotRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("NotFnRune", func() Update[T] { return newFnRune[T](b.logger, Not(pred), FnAccept) })
	return
}

func (b Builder[T]) FollowedByCheckNotRune(pred RunePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FollowedByCheckNotRune", func() Update[T] { return newFnRune[T](b.logger, Not(pred), FnLook) })
	return
}

// Rune is a state that matches the given rune.
func (b Builder[T]) Rune(sample rune) (tail *Chain[T]) {
	tail = b.append("Rune", func() Update[T] { return newFnRune[T](b.logger, IsRune(sample), FnAccept) })
	return
}

func (b Builder[T]) FollowedByRune(sample rune) (tail *Chain[T]) {
	tail = b.append("FollowedByRune", func() Update[T] { return newFnRune[T](b.logger, IsRune(sample), FnLook) })
	return
}

// NotRune is a state that matches all runes except the given one.
func (b Builder[T]) NotRune(sample rune) (tail *Chain[T]) {
	tail = b.append("NotRune", func() Update[T] { return newFnRune[T](b.logger, Not(IsRune(sample)), FnAccept) })
	return
}

func (b Builder[T]) FollowedByNotRune(sample rune) (tail *Chain[T]) {
	tail = b.append("FollowedByNotRune", func() Update[T] { return newFnRune[T](b.logger, Not(IsRune(sample)), FnLook) })
	return
}

// AnyRune is a state that matches any rune.
func (b Builder[T]) AnyRune() (tail *Chain[T]) {
	tail = b.append("AnyRune", func() Update[T] { return newFnRune[T](b.logger, True[rune](), FnAccept) })
	return
}

func (b Builder[T]) FollowedByAnyRune() (tail *Chain[T]) {
	tail = b.append("FollowedByAnyRune", func() Update[T] { return newFnRune[T](b.logger, True[rune](), FnLook) })
	return
}
