package state

import (
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

type (
	// String is a state that matches the given string.
	String struct {
		logger common.Logger
		sample StringSampleProvider
		pred   StringPredicate
	}

	// StringSampleProvider is a function that returns the sample string to match.
	StringSampleProvider func() string
	// StringPredicate is a function that checks if the given string matches the sample.
	StringPredicate func(string, string) bool
)

func newString[T any](
	logger common.Logger,
	sample StringSampleProvider,
	pred StringPredicate,
) *String {
	return &String{
		logger: logger,
		sample: sample,
		pred:   pred,
	}
}

// Update implements State interface.
func (s String) Update(ctx context.Context, tx xio.State) (err error) {
	sample := s.sample()
	size := len(sample)
	buffer := make([]byte, size)
	n, err := tx.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	in := buffer[:n]
	if !s.pred(sample, string(in)) {
		err = ErrRollback
		return
	}
	err = ErrNext
	return
}

func (b Builder[T]) stringState(
	name string,
	source StringSampleProvider,
	pred StringPredicate,
) (tail *Chain[T]) {
	tail = b.createNode(name, func() any {
		return newString[T](
			b.logger,
			source,
			pred,
		)
	})
	return
}

// String is a state that compares the given sample with state input.
// It will has positive result if sample is equal to state input.
func (b Builder[T]) String(sample string) (tail *Chain[T]) {
	tail = b.stringState("String", func() string { return sample }, func(sample, in string) bool { return sample == in })
	return
}

// NotString is a state that compares the given sample with state input.
// It will has positive result if sample is not equal to state input.
func (b Builder[T]) NotString(sample string) (tail *Chain[T]) {
	tail = b.stringState("NotString", func() string { return sample }, func(sample, in string) bool { return sample != in })
	return
}
