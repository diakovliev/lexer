package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/message"
)

// Builder is a builder for the chains of states.
type Builder[T any] struct {
	logger   common.Logger
	next     *Chain[T]
	receiver message.Receiver[T]
}

// Make creates a new builder for the chains of states.
func Make[T any](logger common.Logger, receiver message.Receiver[T]) Builder[T] {
	return Builder[T]{
		logger:   logger,
		receiver: receiver,
	}
}

// createNode creates a new node in chain of states and returns the head of the chain.
func (b Builder[T]) createNode(name string, newState func() any) (head *Chain[T]) {
	created := newState()
	var state State[T]
	var ok bool
	if state, ok = created.(State[T]); !ok {
		b.logger.Fatal("state must implement State[T] interface")
		return
	}
	node := newChain(b, name, state)
	if node.Builder.next != nil {
		node.Builder.next.Append(node)
	} else {
		node.Builder.next = node
	}
	head = node.Builder.next
	node.head = head
	return
}
