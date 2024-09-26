package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/xio"
)

type (
	fixResultState[T any] struct {
		Result error
	}
)

func (t fixResultState[T]) Update(_ context.Context, _ xio.State) error {
	return t.Result
}

func newFixedResultState[T any](result error) func() Update[T] {
	return func() Update[T] {
		return &fixResultState[T]{
			Result: result,
		}
	}
}

func newFakeState[T any]() Update[T] {
	return &fixResultState[T]{
		Result: errors.New("test error"),
	}
}
