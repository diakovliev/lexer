package states

import (
	"errors"

	"github.com/diakovliev/lexer/common"
)

type (
	Chain[T any] struct {
		Builder[T]
		logger   common.Logger
		state    State[T]
		name     string
		messages []common.Message[T]
		head     *Chain[T]
		next     *Chain[T]
	}
)

// newChain creates a new instance of Node struct. It takes logger and name as parameters.
// The name parameter is used for logging purposes.
func newChain[T any](factory Builder[T], name string, state State[T]) *Chain[T] {
	return &Chain[T]{
		Builder: factory,
		logger:  factory.logger,
		name:    name,
		state:   state,
	}
}

// receiver receives a message and stores it in messages slice.
func (c *Chain[T]) receiver(m common.Message[T]) (err error) {
	c.messages = append(c.messages, m)
	return
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

// commitMessages sends all messages in the chain head messages slice to the Builder receiver.
func (c *Chain[T]) commitMessages() (err error) {
	for _, m := range c.head.messages {
		if sendErr := c.head.Builder.receiver(m); sendErr != nil {
			c.logger.Error("send message error: %s", m)
			err = sendErr
			return
		}
	}
	c.head.messages = c.head.messages[:0]
	return
}

// Update implements State interface
func (c *Chain[T]) Update(tx common.ReadUnreadData) (err error) {
	c.logger.Trace("=>> enter Update()")
	defer c.logger.Trace("<<= leave Update() = err=%s", err)
	currentNode := c
	for currentNode != nil {
		c.logger.Trace("%s.Update()", currentNode.name)
		if err = currentNode.state.Update(tx); err == nil {
			c.logger.Fatal("%s.Update() = nil", currentNode.name)
		}
		c.logger.Trace("%s.Update() = err=%s", currentNode.name, err)
		switch {
		case errors.Is(err, errChainNext) || (errors.Is(err, ErrCommit) && currentNode.Next() != nil):
			if errors.Is(err, ErrCommit) {
				c.logger.Trace("commit messages")
				if commitMessagesErr := c.commitMessages(); commitMessagesErr != nil {
					c.logger.Error("commit messages error: %s", commitMessagesErr)
					err = commitMessagesErr
					return
				}
			}
			c.logger.Trace("continue chain")
			currentNode = currentNode.Next()
		case errors.Is(err, ErrCommit):
			c.logger.Trace("commit messages")
			if commitMessagesErr := c.commitMessages(); commitMessagesErr != nil {
				c.logger.Error("commit messages error: %s", commitMessagesErr)
				err = commitMessagesErr
			}
			return
		case errors.Is(err, ErrRollback):
			c.logger.Trace("rollback")
			return
		case errors.Is(err, errBreak):
			c.logger.Trace("break")
			return
		default:
			c.logger.Error("unexpected error: %v", err)
			return
		}
	}
	return
}
