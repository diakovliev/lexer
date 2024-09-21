package state

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
)

// String is a state that matches the given string.
type String struct {
	logger common.Logger
	input  string
}

func newString[T any](input string, logger common.Logger) *String {
	return &String{
		logger: logger,
		input:  input,
	}
}

// Update implements State interface.
func (s String) Update(ctx context.Context, tx xio.State) (err error) {
	size := len(s.input)
	buffer := bytes.NewBuffer(nil)
	buffer.Grow(size)
	n, err := io.CopyN(buffer, tx, int64(size))
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if int(n) != len(s.input) || buffer.String() != s.input {
		err = ErrRollback
		return
	}
	err = ErrNext
	return
}

// String is a state that matches the given string.
func (b Builder[T]) String(input string) (tail *Chain[T]) {
	defaultName := "String"
	tail = b.createNode(defaultName, func() any { return newString[T](input, b.logger) })
	return
}
