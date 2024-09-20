package state

import (
	"errors"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/xio"
)

type (
	Chain[T any] struct {
		Builder[T]
		Receiver *message.SliceReceiver[T]
		logger   common.Logger
		state    State[T]
		name     string
		head     *Chain[T]
		next     *Chain[T]
	}
)

// newChain creates a new instance of Node struct. It takes logger and name as parameters.
// The name parameter is used for logging purposes.
func newChain[T any](factory Builder[T], name string, state State[T]) *Chain[T] {
	return &Chain[T]{
		Builder:  factory,
		Receiver: message.Slice[T](),
		logger:   factory.logger,
		name:     name,
		state:    state,
	}
}

// Next returns the next state in the chain of nodes. If there is no next nodes, it returns an nil.
func (c *Chain[T]) Next() *Chain[T] {
	return c.next
}

// Last returns the last state in the chain of nodes. If there is no next node, it returns current node.
func (c *Chain[T]) Last() *Chain[T] {
	currentState := c
	for currentState.Next() != nil {
		currentState = currentState.Next()
	}
	return currentState
}

// Append appends a new state to the end of the chain of nodes. If there is no next node,
// it sets the next node as the one passed in parameter. Otherwise, it updates the last node's next
// field with the one passed in parameter. It returns the updated node.
func (c *Chain[T]) Append(node *Chain[T]) *Chain[T] {
	if c.next == nil {
		c.next = node
		return c
	}
	last := c.Last()
	last.next = node
	return c
}

// Update implements State interface
func (c *Chain[T]) Update(tx xio.State) (err error) {
	c.logger.Trace("=>> enter Update()")
	defer func() { c.logger.Trace("<<= leave Update() = err=%s", err) }()
	current := c
	for current != nil {
		c.logger.Trace("%s.Update()", current.name)
		if err = current.state.Update(tx); err == nil {
			c.logger.Fatal("%s.Update() = nil", current.name)
		}
		c.logger.Trace("%s.Update() = err=%s", current.name, err)
		next := current.Next()
		switch {
		case errors.Is(err, ErrNext):
			// nothing to do, just move on to the next state
		case errors.Is(err, ErrCommit):
			c.logger.Trace("commit messages")
			if emitToErr := c.head.Receiver.EmitTo(c.head.Builder.Receiver); emitToErr != nil {
				c.logger.Fatal("emit to error: %s", emitToErr)
			}
			if next == nil {
				return
			}
		case errors.Is(err, ErrRollback):
			c.logger.Trace("rollback")
			return
		case errors.Is(err, ErrBreak):
			c.logger.Trace("break")
			return
		default:
			c.logger.Error("unexpected error: %v", err)
			return
		}
		c.logger.Trace("continue chain")
		current = next
	}
	return
}
