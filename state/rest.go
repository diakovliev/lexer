package state

import (
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	Rest[T any] struct {
		logger common.Logger
	}
)

func newRest[T any](logger common.Logger) *Rest[T] {
	return &Rest[T]{
		logger: logger,
	}
}

func (r *Rest[T]) Update(tx xio.State) (err error) {
	// just advance the reader and do nothing else
	_, _ = io.Copy(io.Discard, tx)
	err = ErrNext
	return
}

func (b Builder[T]) Rest() (head *Chain[T]) {
	defaultName := "Rest"
	head = b.createNode(defaultName, func() any { return newRest[T](b.logger) })
	return
}
