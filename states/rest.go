package states

import (
	"io"

	"github.com/diakovliev/lexer/common"
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

func (r *Rest[T]) Update(tx common.ReadUnreadData) (err error) {
	// just advance the reader and do nothing else
	_, _ = io.Copy(io.Discard, tx)
	err = errChainNext
	return
}

func (b Builder[T]) Rest() (head *Chain[T]) {
	defaultName := "Rest"
	head = b.createNode(defaultName, func() any { return newRest[T](b.logger) })
	return
}
