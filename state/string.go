package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// String is a state that matches the given string.
type String struct {
	logger common.Logger
	sample func() string
	pred   func(string) bool
}

func newString[T any](logger common.Logger, sample func() string, pred func(string) bool) *String {
	return &String{
		logger: logger,
		sample: sample,
		pred:   pred,
	}
}

// Update implements State interface.
func (s String) Update(ctx context.Context, tx xio.State) (err error) {
	size := len(s.sample())
	buffer := make([]byte, size)
	n, err := tx.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	in := buffer[:n]
	if !s.pred(string(in)) {
		err = ErrRollback
		return
	}
	err = ErrNext
	return
}

// String is a state that compares the given sample with state input.
// It will has positive result if sample is equal to state input.
func (b Builder[T]) String(sample string) (tail *Chain[T]) {
	defaultName := "String"
	tail = b.createNode(defaultName, func() any {
		return newString[T](
			b.logger,
			func() string { return sample },
			func(s string) bool { return sample == s },
		)
	})
	return
}

// NotString is a state that compares the given sample with state input.
// It will has positive result if sample is not equal to state input.
func (b Builder[T]) NotString(sample string) (tail *Chain[T]) {
	defaultName := "NotString"
	tail = b.createNode(defaultName, func() any {
		return newString[T](
			b.logger,
			func() string { return sample },
			func(s string) bool { return sample != s },
		)
	})
	return
}
