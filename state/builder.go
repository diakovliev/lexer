package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
)

// Builder is a builder.
type Builder[T any] struct {
	factory  message.Factory[T]
	receiver message.Receiver[T]
	logger   common.Logger
	last     *Chain[T]
}

// Make creates a new builder.
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

// append creates a new node in chain of states and returns the tail of the chain.
// if builder is not yet associated with any chain then it creates the new one and
// returns its tail. Each element in the chain is a builder.
func (b Builder[T]) append(stateName string, newState func() any) (tail *Chain[T]) {
	prev := b.last
	if prev != nil && prev.next != nil {
		b.logger.Fatal("invalid grammar: last element already has next: %T", prev.next)
	}
	created := newState()
	var state Update[T]
	var ok bool
	if state, ok = created.(Update[T]); !ok {
		b.logger.Fatal("not a state: %T", created)
	}
	var node *Chain[T]
	if prev == nil {
		tail = newChain(b, stateName, state)
		tail.Builder.last = tail
		return
	}
	node = newChain(prev.Builder, prev.name+"."+stateName, state)
	node.Builder.last = node
	node.prev = prev
	prev.next = node
	tail = node
	return
}
