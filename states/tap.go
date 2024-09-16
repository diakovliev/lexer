package states

import "github.com/diakovliev/lexer/common"

type Tap[T any] struct {
	logger common.Logger
	fn     func() error
}

func newTap[T any](logger common.Logger, fn func() error) *Tap[T] {
	return &Tap[T]{
		logger: logger,
		fn:     fn,
	}
}

func (t Tap[T]) Update(_ common.ReadUnreadData) (err error) {
	if err = t.fn(); err != nil {
		return
	}
	// return ErrCommit to be able to Tap end of the chain
	err = ErrCommit
	return
}

func (b Builder[T]) Tap(fn func() error) (head *Chain[T]) {
	defaultName := "Tap"
	head = b.createNode(defaultName, func() any { return newTap[T](b.logger, fn) })
	return
}
