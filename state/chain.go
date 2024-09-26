package state

import (
	"context"
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	Chain[T any] struct {
		Builder[T]
		logger   common.Logger
		name     string
		prev     *Chain[T]
		next     *Chain[T]
		state    Update[T]
		receiver *message.SliceReceiver[T]
	}
)

// newChain creates a new instance of Node struct. It takes logger and name as parameters.
// The name parameter is used for logging purposes.
func newChain[T any](builder Builder[T], name string, state Update[T], createReceiver bool) (ret *Chain[T]) {
	ret = &Chain[T]{
		Builder: builder,
		name:    name,
		logger:  builder.logger,
		state:   state,
	}
	if createReceiver {
		ret.receiver = message.Slice[T]()
	}
	return
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

// ForwardMessages forwards messages from the head to final receiver.
func (c *Chain[T]) ForwardMessages() (err error) {
	head := c.Head()
	if err = head.receiver.ForwardTo(head.Builder.receiver); err != nil {
		err = fmt.Errorf("%s: %w", head.name, err)
	}
	return
}

// Update implements State interface
func (c *Chain[T]) Update(ctx context.Context, ioState xio.State) (err error) {
	current := c.Head()
	for current != nil {
		next := current.Next()
		if err = current.state.Update(withStateName(ctx, current.name), ioState); err == nil {
			c.logger.Fatal("unexpected nil")
		}
		if errors.Is(err, errChainRepeat) {
			prev := current.Prev()
			if prev == nil {
				c.logger.Fatal("unexpected nil")
				return
			}
			err = c.repeat(withStateName(ctx, prev.name), prev.state, err, ioState)
		}
		switch {
		case errors.Is(err, errChainNext):
			if next == nil {
				c.logger.Fatal("invalid grammar: next can't be from last in chain")
			}
			if next != nil && isZeroMaxRepeat[T](next.state) {
				err = ErrRollback
				return
			}
		case errors.Is(err, ErrCommit):
			if next != nil && isZeroMaxRepeat[T](next.state) {
				err = ErrRollback
				return
			}
			if err := c.ForwardMessages(); err != nil {
				c.logger.Fatal("forward messages error: %s", err)
			}
			if next == nil {
				return
			}
		case errors.Is(err, ErrRollback):
			if next == nil {
				return
			}
			if !isZeroMinRepeat[T](next.state) {
				return
			}
			if isRepeat[T](current.state) {
				return
			}
			err = errChainNext
		case errors.Is(err, errStateBreak):
			if next != nil {
				c.logger.Fatal("invalid grammar: break must be last in chain")
			}
			if err := c.ForwardMessages(); err != nil {
				c.logger.Fatal("forward messages error: %s", err)
			}
			return
		case errors.Is(err, ErrIncomplete), errors.Is(err, ErrInvalidInput):
			// pass known errors as is
		default:
			// wrap all other errors with state break error
			err = makeErrBreak(err)
			return
		}
		current = next
	}
	return
}
