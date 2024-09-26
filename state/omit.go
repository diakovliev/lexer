package state

import (
	"context"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// Omit is a state what omits current input data, without producing message.
type Omit[T any] struct {
	logger common.Logger
}

// newOmit creates new Omit state.
func newOmit[T any](logger common.Logger) *Omit[T] {
	return &Omit[T]{
		logger: logger,
	}
}

// Update implements Update interface
func (o Omit[T]) Update(ctx context.Context, ioState xio.State) (err error) {
	data, _, err := ioState.Data()
	common.AssertNoError(err, "data error")
	common.AssertFalse(len(data) == 0, "nothing to omit")
	err = ErrCommit
	return
}

// Omit adds omit state to the chain.
// It omits current input data, without producing message.
// If there are no input data, it panics.
func (b Builder[T]) Omit() (tail *Chain[T]) {
	common.AssertNotNilPtr(b.last, "invalid grammar: omit can't be the first state in chain")
	tail = b.append("Omit", func() Update[T] { return newOmit[T](b.logger) })
	return
}

// isOmit checks if state is Omit.
func isOmit[T any](s Update[T]) (ret bool) {
	_, ret = s.(*Omit[T])
	return
}
