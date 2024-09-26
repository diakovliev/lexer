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
	mode   FnMode
}

// newFnRune creates a new state that checks if the next rune matches the predicate.
func newFnByte[T any](logger common.Logger, pred BytePredicate, mode FnMode) *FnByte[T] {
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
	case FnAccept:
		if !result {
			if _, unreadErr := tx.Unread(); unreadErr != nil {
				fb.logger.Fatal("unread error: %s", unreadErr)
			}
			err = ErrRollback
			return
		}
	case FnLook:
		if _, unreadErr := tx.Unread(); unreadErr != nil {
			fb.logger.Fatal("unread error: %s", unreadErr)
		}
	}
	if result {
		err = errChainNext
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
	ok = i.mode == FnLook
	return ok
}

// CheckByte adds a state that checks if the next rune matches the predicate to the chain.
func (b Builder[T]) CheckByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FnByte", func() Update[T] { return newFnByte[T](b.logger, pred, FnAccept) })
	return
}

func (b Builder[T]) FollowedByCheckByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FollowedByCheckByte", func() Update[T] { return newFnByte[T](b.logger, pred, FnLook) })
	return
}

// CheckNotByte adds a state that checks if the next rune doesn't match the predicate to the chain.
func (b Builder[T]) CheckNotByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("NotFnByte", func() Update[T] { return newFnByte[T](b.logger, Not(pred), FnAccept) })
	return
}

func (b Builder[T]) FollowedByCheckNotByte(pred BytePredicate) (tail *Chain[T]) {
	if pred == nil {
		b.logger.Fatal("invalid grammar: nil predicate")
	}
	tail = b.append("FollowedByCheckNotByte", func() Update[T] { return newFnByte[T](b.logger, Not(pred), FnLook) })
	return
}

// Byte adds a state that checks if the next rune matches the sample to the chain.
func (b Builder[T]) Byte(sample byte) (tail *Chain[T]) {
	tail = b.append("Byte", func() Update[T] { return newFnByte[T](b.logger, IsByte(sample), FnAccept) })
	return
}

func (b Builder[T]) FollowedByByte(sample byte) (tail *Chain[T]) {
	tail = b.append("FollowedByByte", func() Update[T] { return newFnByte[T](b.logger, IsByte(sample), FnLook) })
	return
}

// NotByte adds a state that checks if the next rune doesn't match the sample to the chain.
func (b Builder[T]) NotByte(sample byte) (tail *Chain[T]) {
	tail = b.append("NotByte", func() Update[T] { return newFnByte[T](b.logger, Not(IsByte(sample)), FnAccept) })
	return
}

func (b Builder[T]) FollowedByNotByte(sample byte) (tail *Chain[T]) {
	tail = b.append("FollowedByNotByte", func() Update[T] { return newFnByte[T](b.logger, Not(IsByte(sample)), FnLook) })
	return
}

// ByteRange adds a state that checks if the next rune matches the range to the chain.
func (b Builder[T]) AnyByte() (tail *Chain[T]) {
	tail = b.append("AnyByte", func() Update[T] { return newFnByte[T](b.logger, True[byte](), FnAccept) })
	return
}

func (b Builder[T]) FollowedByAnyByte() (tail *Chain[T]) {
	tail = b.append("FollowedByAnyByte", func() Update[T] { return newFnByte[T](b.logger, True[byte](), FnLook) })
	return
}
