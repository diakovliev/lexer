package vm

import (
	"testing"

	"slices"

	"github.com/stretchr/testify/assert"
)

func TestVm1(t *testing.T) {
	code := []Cell{
		{Op: Add},
		{Op: Val, Value: int64(2)},
		{Op: Val, Value: int64(2)},
		{Op: Val, Value: int64(2)},
	}
	slices.Reverse(code)
	vm := New().PushCode(code)
	err := vm.Run()
	assert.ErrorIs(t, err, ErrHalt)
	cell, err := vm.Peek()
	assert.NoError(t, err)
	assert.Len(t, vm.stack, 1)
	assert.Equal(t, Cell{Op: Val, Value: int64(2)}, cell)
}
