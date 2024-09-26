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
func (b Builder[T]) append(name string, newState func() Update[T]) (tail *Chain[T]) {
	state := newState()
	common.AssertNotNil(state, "nil state")
	if b.last == nil {
		// new chain
		tail = newChain(b, name, state, b.last)
		return
	}
	// append to existing chain
	common.AssertNilPtr(b.last.next(), "invalid grammar: last element already has next")
	tail = newChain(b.last.Builder, b.last.name()+"."+name, state, b.last)
	return
}
