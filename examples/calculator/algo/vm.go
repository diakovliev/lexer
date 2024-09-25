package algo

import (
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
)

var (
	// ErrVMHalt is returned when the VM halted (stack is empty)
	ErrVMHalt = errors.New("halt")
	// ErrVMNonOperation is returned when the VM received non operation token when it was expected to receive operation token.
	ErrVMNonOperation = errors.New("non operation")
	// ErrVMUnknownOperation is returned when the VM received unknown operation token.
	ErrVMUnknownOperation = errors.New("unknown operation")
	// ErrVMStackEmpty is returned by vm.Pop() when the VM stack is empty.
	ErrVMStackEmpty = errors.New("stack is empty")
)

type (
	// VM is a simple stack virtual machine
	VM struct {
		stack stack[VMCode]
	}

	// VMCode is a code of the virtual machine. It contains token and its value.
	VMCode struct {
		Token grammar.Token
		Value int
	}
)

// NewVM creates new virtual machine with given code.
func NewVM(code []VMCode) (vm *VM) {
	vm = &VM{
		stack: makeStack[VMCode](len(code)),
	}
	for _, token := range code {
		vm.stack = vm.stack.Push(token)
	}
	return
}

// Push pushes new token to the stack of virtual machine.
func (vm *VM) Push(t VMCode) {
	vm.stack = vm.stack.Push(t)
}

// Pop pops token from the stack of virtual machine and returns it.
// If there is no tokens in the stack, then error will be returned.
func (vm *VM) Pop() (value VMCode, err error) {
	if vm.stack.Empty() {
		err = ErrVMStackEmpty
		return
	}
	vm.stack, value = vm.stack.Pop()
	return
}

func (vm *VM) fetchCommand() (cmd VMCode, err error) {
	if vm.stack.Empty() {
		err = ErrVMStackEmpty
		return
	}
	vm.stack, cmd = vm.stack.Pop()
	if !Ops.HasToken(cmd.Token) {
		err = ErrVMNonOperation
	}
	return
}

func (vm *VM) fetch() (token VMCode, err error) {
	if vm.stack.Empty() {
		err = ErrVMStackEmpty
		return
	}
	vm.stack, token = vm.stack.Pop()
	if Ops.HasToken(token.Token) {
		vm.stack = vm.stack.Push(token)
		if err = vm.step(); err != nil && !errors.Is(err, ErrVMHalt) {
			return
		}
		vm.stack, token = vm.stack.Pop()
		err = nil
	}
	return
}

func (vm *VM) execute(cmd VMCode) (err error) {
	// TODO: how many operands is needed?
	var left, right VMCode
	if right, err = vm.fetch(); err != nil {
		return
	}
	if left, err = vm.fetch(); err != nil {
		return
	}
	var result int
	switch cmd.Token {
	case grammar.Plus:
		result = left.Value + right.Value
	case grammar.Minus:
		result = left.Value - right.Value
	case grammar.Mul:
		result = left.Value * right.Value
	case grammar.Div:
		result = left.Value / right.Value
	default:
		err = fmt.Errorf("%w: %d", ErrVMUnknownOperation, cmd)
	}
	if vm.stack.Empty() {
		err = ErrVMHalt
	}
	// push result
	vm.stack = vm.stack.Push(VMCode{
		Token: grammar.Number,
		Value: result,
	})
	return
}

func (vm *VM) step() (err error) {
	if vm.stack.Empty() {
		err = ErrVMStackEmpty
		return
	}
	cmd, err := vm.fetchCommand()
	if err != nil {
		return
	}
	err = vm.execute(cmd)
	return
}

// Run the VM, return ErrVMHalt when finished.
func (vm *VM) Run() (err error) {
	for err = vm.step(); err == nil; {
	}
	return
}
