package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Names is a pseudo state that is used to mark the begin of a named chain.
type Named[T any] struct {
	logger common.Logger
}

// newNamed creates a new instance of the named state.
func newNamed[T any](logger common.Logger) *Named[T] {
	return &Named[T]{
		logger: logger,
	}
}

// Update implements Update interface
func (n Named[T]) Update(_ context.Context, _ xio.State) (err error) {
	err = errChainNext
	return
}

// Named creates a named chain. It must be the first state in the chain.
func (b Builder[T]) Named(name string) (tail *Chain[T]) {
	if b.last != nil {
		b.logger.Fatal("invalid grammar: named must be the first state in the chain")
	}
	tail = b.append(name, func() any { return newNamed[T](b.logger) })
	return
}

// isNamed checks if the state is a named state.
func isNamed[T any](s Update[T]) (ok bool) {
	_, ok = s.(Named[any])
	return
}
