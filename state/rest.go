package state

import (
	"context"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Rest is a state that consumes all the remaining input from the io state.
type Rest[T any] struct {
	logger common.Logger
}

// newRest creates a new Rest state.
func newRest[T any](logger common.Logger) *Rest[T] {
	return &Rest[T]{
		logger: logger,
	}
}

// Update implements the State interface. It consumes all the remaining input from the io state.
func (r *Rest[T]) Update(ctx context.Context, ioState xio.State) (err error) {
	// just advance the reader and do nothing else
	_, _ = io.Copy(io.Discard, ioState)
	err = ErrChainNext
	return
}

// Rest adds a Rest state to the chain. It consumes all the remaining input from the io state.
func (b Builder[T]) Rest() (tail *Chain[T]) {
	tail = b.append("Rest", func() Update[T] { return newRest[T](b.logger) })
	return
}

// IsRest checks if the state is a Rest state.
func isRest[T any](s Update[T]) bool {
	_, ok := s.(*Rest[T])
	return ok
}
