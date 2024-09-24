package algo

import (
	"errors"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
)

var (
	ErrVmComplete = errors.New("vm complete")
)

// Simple stack virtual machine
type Vm struct {
	stack stack[VmData]
}

func NewVm(code []VmData) (vm *Vm) {
	vm = &Vm{}
	for _, token := range code {
		vm.stack = vm.stack.Push(token)
	}
	return
}

func (vm *Vm) Push(t VmData) {
	vm.stack = vm.stack.Push(t)
}

func (vm *Vm) Pop() (value VmData) {
	vm.stack, value = vm.stack.Pop()
	return value
}

func (vm *Vm) getOperand() (oper VmData, err error) {
	vm.stack, oper = vm.stack.Pop()
	if Ops.HasToken(oper.Token) {
		vm.stack = vm.stack.Push(oper)
		if err = vm.step(); err != nil && !errors.Is(err, ErrVmComplete) {
			return
		}
		vm.stack, oper = vm.stack.Pop()
		err = nil
	}
	return
}

func (vm *Vm) step() (err error) {
	// pop operator
	var token VmData
	vm.stack, token = vm.stack.Pop()
	if !Ops.HasToken(token.Token) {
		err = errors.New("unexpected token")
		return
	}
	var operL, operR VmData
	if operR, err = vm.getOperand(); err != nil {
		return
	}
	if operL, err = vm.getOperand(); err != nil {
		return
	}
	// calculate
	var result int
	switch token.Token {
	case grammar.Plus:
		result = operL.Value + operR.Value
	case grammar.Minus:
		result = operL.Value - operR.Value
	case grammar.Mul:
		result = operL.Value * operR.Value
	case grammar.Div:
		result = operL.Value / operR.Value
	default:
		err = errors.New("unexpected token")
	}
	if vm.stack.Empty() {
		err = ErrVmComplete
	}
	// push result
	vm.stack = vm.stack.Push(VmData{
		Token: grammar.Number,
		Value: result,
	})
	return
}

func (vm *Vm) Execute() (err error) {
	for err = vm.step(); err == nil; {
	}
	return
}
