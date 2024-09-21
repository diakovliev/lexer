package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
)

// Builder is a builder for the chains of states.
type Builder[T any] struct {
	factory  message.Factory[T]
	receiver message.Receiver[T]
	logger   common.Logger
	last     *Chain[T]
}

// Make creates a new builder for the chains of states.
func Make[T any](
	logger common.Logger,
	factory message.Factory[T],
	receiver message.Receiver[T],
) Builder[T] {
	return Builder[T]{
		logger:   logger,
		factory:  factory,
		receiver: receiver,
	}
}

// createNode creates a new node in chain of states and returns the head of the chain.
func (b Builder[T]) createNode(name string, newState func() any) (tail *Chain[T]) {
	created := newState()
	var state State[T]
	var ok bool
	if state, ok = created.(State[T]); !ok {
		b.logger.Fatal("not a state: %T", created)
	}
	prev := b.last
	var node *Chain[T]
	if prev == nil {
		tail = newChain(b, name, state)
		tail.Builder.last = tail
		return
	}
	node = newChain(prev.Builder, name, state)
	node.Builder.last = node
	node.prev = prev
	prev.next = node
	tail = node
	return
}
