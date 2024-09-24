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

func NewVm() *Vm {
	return &Vm{}
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
	var oper1, oper2 VmData
	if oper1, err = vm.getOperand(); err != nil && !errors.Is(err, ErrVmComplete) {
		return
	}
	if oper2, err = vm.getOperand(); err != nil && !errors.Is(err, ErrVmComplete) {
		return
	}
	// calculate
	var result int
	switch token.Token {
	case grammar.Plus:
		result = oper1.Value + oper2.Value
	case grammar.Minus:
		result = oper1.Value - oper2.Value
	case grammar.Mul:
		result = oper1.Value * oper2.Value
	case grammar.Div:
		result = oper1.Value / oper2.Value
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

func (vm *Vm) Execute(bnf []VmData) (err error) {
	for _, token := range bnf {
		vm.stack = vm.stack.Push(token)
	}
	for {
		err = vm.step()
		if err != nil {
			return
		}
	}
}
