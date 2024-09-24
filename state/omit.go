package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type Omit[T any] struct {
	logger common.Logger
}

func newOmit[T any](logger common.Logger) *Omit[T] {
	return &Omit[T]{
		logger: logger,
	}
}

func (o Omit[T]) Update(ctx context.Context, tx xio.State) (err error) {
	data, _, err := tx.Data()
	if err != nil {
		o.logger.Fatal("data error: %s", err)
	}
	if len(data) == 0 {
		o.logger.Fatal("nothing to omit")
	}
	err = errCommit
	return
}

func (b Builder[T]) Omit() (tail *Chain[T]) {
	if b.last == nil {
		b.logger.Fatal("invalid grammar: omit can't be the first state in chain")
	}
	tail = b.append("Omit", func() any { return newOmit[T](b.logger) })
	return
}

func isOmit[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Omit[T])
	return
}
