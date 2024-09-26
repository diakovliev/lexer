package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	// FixedResult is a pseudo state that will always return given error.
	FixedResult[T any] struct {
		logger common.Logger
		err    error
	}

	// Named is a pseudo state that will always return given error.
	Named[T any] struct {
		*FixedResult[T]
	}
)

// NewFixedResult creates a new instance of the fixed result state.
func newFixedResult[T any](logger common.Logger, err error) *FixedResult[T] {
	return &FixedResult[T]{
		logger: logger,
		err:    err,
	}
}

// Update implements Update interface
func (fr FixedResult[T]) Update(_ context.Context, _ xio.State) (err error) {
	err = fr.err
	return
}

// newNamed creates a new instance of the named state.
func newNamed[T any](logger common.Logger) *Named[T] {
	return &Named[T]{
		FixedResult: newFixedResult[T](logger, errChainNext),
	}
}

// Named creates a named chain. It must be the first state in the chain.
func (b Builder[T]) Named(name string) (tail *Chain[T]) {
	common.AssertNilPtr(b.last, "invalid grammar: named must be the first state in the chain")
	tail = b.append(name, func() Update[T] { return newNamed[T](b.logger) })
	return
}

// isNamed checks if the state is a named state.
func isNamed[T any](s Update[T]) (ok bool) {
	_, ok = s.(Named[any])
	return
}
