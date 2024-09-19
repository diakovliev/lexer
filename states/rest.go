package states

import (
	"io"

	"github.com/diakovliev/lexer/common"
)

type (
	Rest[T any] struct {
		logger common.Logger
	}
	restDiscard struct{}
)

var (
	restBufferSize = 10
)

func (restDiscard) Write(p []byte) (int, error) {
	return len(p), nil
}

func newRest[T any](logger common.Logger) *Rest[T] {
	return &Rest[T]{
		logger: logger,
	}
}

func (r *Rest[T]) Update(tx common.ReadUnreadData) (err error) {
	// just advance the reader and do nothing else
	//_, _ = io.CopyBuffer(io.Discard, tx, make([]byte, restBufferSize))
	_, _ = io.CopyBuffer(restDiscard{}, tx, make([]byte, restBufferSize))
	err = errChainNext
	return
}

func (b Builder[T]) Rest() (head *Chain[T]) {
	defaultName := "Rest"
	head = b.createNode(defaultName, func() any { return newRest[T](b.logger) })
	return
}
