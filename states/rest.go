package states

import (
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
)

type Rest[T any] struct {
	logger common.Logger
}

func newRest[T any](logger common.Logger) *Rest[T] {
	return &Rest[T]{
		logger: logger,
	}
}

func (r *Rest[T]) Update(tx common.ReadUnreadData) (err error) {
	data := make([]byte, 25)
	for {
		_, err = tx.Read(data)
		if errors.Is(err, io.EOF) {
			break
		}
	}
	err = errChainNext
	return
}

func (b Builder[T]) Rest() (head *Chain[T]) {
	defaultName := "Rest"
	head = b.createNode(defaultName, func() any { return newRest[T](b.logger) })
	return
}
