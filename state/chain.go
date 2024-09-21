package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	Chain[T any] struct {
		Builder[T]
		logger   common.Logger
		prev     *Chain[T]
		next     *Chain[T]
		state    State[T]
		name     string
		receiver *message.SliceReceiver[T]
	}
)

// newChain creates a new instance of Node struct. It takes logger and name as parameters.
// The name parameter is used for logging purposes.
func newChain[T any](builder Builder[T], name string, state State[T]) *Chain[T] {
	return &Chain[T]{
		Builder:  builder,
		receiver: message.Slice[T](),
		logger:   builder.logger,
		name:     name,
		state:    state,
	}
}

// Next returns the next state in the chain of nodes. If there is no next nodes, it returns an nil.
func (c *Chain[T]) Next() *Chain[T] {
	return c.next
}

func (c *Chain[T]) Prev() *Chain[T] {
	return c.prev
}

// Tail returns the last state in the chain of nodes. If there is no next node, it returns current node.
func (c *Chain[T]) Tail() *Chain[T] {
	current := c
	for current.Next() != nil {
		current = current.Next()
	}
	return current
}

func (c *Chain[T]) Head() *Chain[T] {
	current := c
	for current.Prev() != nil {
		current = current.Prev()
	}
	return current
}

// Update implements State interface
func (c *Chain[T]) Update(ctx context.Context, tx xio.State) (err error) {
	head := c.Head()
	current := head
	for current != nil {
		next := current.Next()
		if err = current.state.Update(ctx, tx); err == nil {
			c.logger.Fatal("unexpected nil")
		}
		if errors.Is(err, ErrRepeat) {
			prev := current.Prev()
			if prev == nil {
				c.logger.Fatal("unexpected nil")
				return
			}
			err = c.repeat(ctx, prev.state, err, tx)
		}
		switch {
		case errors.Is(err, ErrNext):
			if next == nil {
				c.logger.Fatal("invalid grammar: next can't be from last in state")
			}
		case errors.Is(err, ErrCommit):
			if err := head.receiver.EmitTo(head.Builder.receiver); err != nil {
				c.logger.Fatal("emit to error: %s", err)
			}
			if next == nil {
				return
			}
		case errors.Is(err, ErrRollback):
			// Repeat(CountBetween(0, N))
			if next == nil {
				return
			}
			if isRepeat[T](current.state) {
				return
			}
			if !isZeroRepeat[T](next.state) {
				return
			}
			err = ErrNext
		case errors.Is(err, ErrBreak):
			if next != nil {
				c.logger.Fatal("invalid grammar: break must be last in state")
			}
			return
		default:
			return
		}
		current = next
	}
	return
}
