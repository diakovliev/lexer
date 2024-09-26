package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	// Chain is a struct that represents a chain of states.
	// It contains a pointer to the previous and next nodes in the chain.
	// Each element in a the chain is a Builder.
	Chain[T any] struct {
		Builder[T]
		p        *Chain[T]
		n        *Chain[T]
		logger   common.Logger
		nodeName string
		ref      Update[T]
		receiver *message.SliceReceiver[T]
	}
)

// newChain creates a new instance of Node struct. It takes logger and name as parameters.
func newChain[T any](
	builder Builder[T],
	name string,
	state Update[T],
	prev *Chain[T],
) (ret *Chain[T]) {
	ret = &Chain[T]{
		Builder:  builder,
		logger:   builder.logger,
		nodeName: name,
		ref:      state,
		p:        prev,
	}
	if ret.p == nil {
		// new chain, create receiver for messages.
		ret.receiver = message.Slice[T]()
	}
	ret.Builder.last = ret
	if ret.p != nil {
		ret.p.n = ret
	}
	return
}

// name returns a name of the node in chain.
func (c *Chain[T]) name() string {
	return c.nodeName
}

// defer returns underlying state.
func (c *Chain[T]) deref() Update[T] {
	return c.ref
}

// next returns the next state in the chain of nodes. If there is no next nodes, it returns an nil.
func (c *Chain[T]) next() *Chain[T] {
	return c.n
}

// prev returns the previous state in the chain of nodes. If there is no previous node, it returns an nil.
func (c *Chain[T]) prev() *Chain[T] {
	return c.p
}

// tail returns the last state in the chain of nodes. If there is no next node, it returns current node.
func (c *Chain[T]) tail() *Chain[T] {
	current := c
	for current.next() != nil {
		current = current.next()
	}
	return current
}

// head returns the first state in the chain of nodes. If there is no previous node, it returns current node.
func (c *Chain[T]) head() *Chain[T] {
	current := c
	for current.prev() != nil {
		current = current.prev()
	}
	return current
}

// forwardMessages forwards messages from the head to final receiver.
func (c *Chain[T]) forwardMessages() {
	head := c.head()
	err := head.receiver.ForwardTo(head.Builder.receiver)
	common.AssertNoError(err, "forward messages error")
}

// Update implements State interface
func (c *Chain[T]) Update(ctx context.Context, ioState xio.State) (err error) {
	current := c.head()
	for current != nil {
		next := current.next()
		err = current.deref().Update(withStateName(ctx, current.name()), ioState)
		common.AssertError(err, "unexpected no error")
		if errors.Is(err, errChainRepeat) {
			prev := current.prev()
			common.AssertNotNilPtr(prev, "no previous state")
			err = c.repeat(withStateName(ctx, prev.name()), prev.deref(), err, ioState)
		}
		switch {
		case errors.Is(err, errChainNext):
			common.AssertNotNilPtr(next, "invalid grammar: next can't be from last in chain")
			if next != nil && isZeroMaxRepeat[T](next.deref()) {
				err = ErrRollback
				return
			}
		case errors.Is(err, ErrCommit):
			if next != nil && isZeroMaxRepeat[T](next.deref()) {
				err = ErrRollback
				return
			}
			c.forwardMessages()
			if next == nil {
				return
			}
		case errors.Is(err, ErrRollback):
			if next == nil {
				return
			}
			if !isZeroMinRepeat[T](next.deref()) {
				return
			}
			if isRepeat[T](current.deref()) {
				return
			}
			err = errChainNext
		case errors.Is(err, errStateBreak):
			common.AssertNilPtr(next, "invalid grammar: next can't be from last in chain")
			c.forwardMessages()
			return
		case errors.Is(err, ErrIncomplete), errors.Is(err, ErrInvalidInput):
			// pass known errors as is
		default:
			// wrap all other errors with state break error
			err = MakeErrBreak(err)
			return
		}
		current = next
	}
	return
}
