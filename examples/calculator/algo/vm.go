package algo

import (
	"errors"
	"fmt"
	"math"
	"slices"

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
	// ErrVMDivByZero is returned by vm.div() when the right value is zero.
	ErrVMDivByZero = errors.New("division by zero")
)

type (
	// VM is a simple stack virtual machine
	VM struct {
		stack stack[VMCode]
	}

	// VMCode is a code of the virtual machine. It contains token and its value.
	VMCode struct {
		Token grammar.Token
		Value any
	}
)

func (vc VMCode) IsZero() bool {
	return vc.AsInt64() == 0 && vc.AsFloat64() == 0
}

func (vc VMCode) IsWhole() bool {
	_, ok := vc.Value.(int64)
	return ok
}

func (vc VMCode) AsInt64() (i int64) {
	i, ok := vc.Value.(int64)
	if !ok {
		f, ok := vc.Value.(float64)
		if !ok {
			panic("not a number")
		}
		i = int64(math.Round(f))
	}
	return
}

func PrintCode(code []VMCode) {
	codeCopy := make([]VMCode, len(code))
	copy(codeCopy, code)
	slices.Reverse(codeCopy)
	for i, c := range codeCopy {
		fmt.Printf("%04d %+v\n", i, c)
	}
}

func (vc VMCode) AsFloat64() (f float64) {
	f, ok := vc.Value.(float64)
	if !ok {
		i, ok := vc.Value.(int64)
		if !ok {
			panic("not a number")
		}
		f = float64(i)
	}
	return
}

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

func (vm *VM) add(left VMCode, right VMCode) (result VMCode) {
	if left.IsWhole() && right.IsWhole() {
		return VMCode{Token: grammar.DecNumber, Value: left.AsInt64() + right.AsInt64()}
	}
	return VMCode{Token: grammar.DecNumber, Value: left.AsFloat64() + right.AsFloat64()}
}

func (vm *VM) sub(left VMCode, right VMCode) (result VMCode) {
	if left.IsWhole() && right.IsWhole() {
		return VMCode{Token: grammar.DecNumber, Value: left.AsInt64() - right.AsInt64()}
	}
	return VMCode{Token: grammar.DecNumber, Value: left.AsFloat64() - right.AsFloat64()}
}

func (vm *VM) mul(left VMCode, right VMCode) (result VMCode) {
	if left.IsWhole() && right.IsWhole() {
		return VMCode{Token: grammar.DecNumber, Value: left.AsInt64() * right.AsInt64()}
	}
	return VMCode{Token: grammar.DecNumber, Value: left.AsFloat64() * right.AsFloat64()}
}

func (vm *VM) div(left VMCode, right VMCode) (result VMCode) {
	return VMCode{Token: grammar.DecNumber, Value: left.AsFloat64() / right.AsFloat64()}
}

func (vm *VM) execute(cmd VMCode) (err error) {
	var left, right VMCode
	if right, err = vm.fetch(); err != nil {
		return
	}
	if left, err = vm.fetch(); err != nil {
		return
	}
	var result VMCode
	switch cmd.Token {
	case grammar.Plus:
		result = vm.add(left, right)
	case grammar.Minus:
		result = vm.sub(left, right)
	case grammar.Mul:
		result = vm.mul(left, right)
	case grammar.Div:
		if right.IsZero() {
			err = ErrVMDivByZero
			return
		}
		result = vm.div(left, right)
	default:
		err = fmt.Errorf("%w: %d", ErrVMUnknownOperation, cmd)
	}
	if vm.stack.Empty() {
		err = ErrVMHalt
	}
	// push result
	vm.stack = vm.stack.Push(result)
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
	PrintCode(vm.stack)
	if len(vm.stack) == 1 && !Ops.HasToken(vm.stack[0].Token) {
		// nothing to do, halt immediately
		err = ErrVMHalt
		return
	}
	for err = vm.step(); err == nil; {
	}
	return
}
