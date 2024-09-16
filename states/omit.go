package states

import "github.com/diakovliev/lexer/common"

type Omit[T any] struct {
	logger common.Logger
}

func newOmit[T any](logger common.Logger) *Omit[T] {
	return &Omit[T]{
		logger: logger,
	}
}

func (o Omit[T]) Update(tx common.ReadUnreadData) (err error) {
	_, data, err := tx.Data()
	if err != nil {
		o.logger.Error("tx.Data() = data=%v, err=%s", data, err)
		return
	}
	if len(data) == 0 {
		o.logger.Fatal("nothing to omit")
		return
	}
	err = ErrCommit
	return
}

func (b Builder[T]) Omit() (head *Chain[T]) {
	defaultName := "Omit"
	head = b.createNode(defaultName, func() any { return newOmit[T](b.logger) })
	return
}
