package vm

import "errors"

var (
	// ErrHalt is returned when the VM halted (stack is empty)
	ErrHalt = errors.New("halt")
	// ErrNonOperation is returned when the VM received non operation token when it was expected to receive operation token.
	ErrNonOperation = errors.New("non operation")
	// ErrUnknownOperation is returned when the VM received unknown operation token.
	ErrUnknownOperation = errors.New("unknown operation")
	// ErrStackEmpty is returned by vm.Pop() when the VM stack is empty.
	ErrStackEmpty = errors.New("stack is empty")
	// ErrDivByZero is returned by vm.div() when the right value is zero.
	ErrDivByZero = errors.New("division by zero")
	// ErrNotEnoughArguments is returned by vm.execute() when the operation has not enough arguments.
	// The on top of the stack operation can't be executed.
	ErrNotEnoughArguments = errors.New("not enough arguments")
)
