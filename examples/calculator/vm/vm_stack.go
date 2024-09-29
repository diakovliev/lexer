package vm

import (
	"fmt"

	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// StackCell
	StackCell struct {
		Op   Cell
		Args []Cell
	}

	VMStackLoop struct {
		vm    *VM
		stack stack.Stack[*StackCell]
	}
)

// AddArg adds new argument to the cell.
func (sc *StackCell) AddArg(c Cell) {
	sc.Args = append(sc.Args, c)
}

// ArgsCount returns the number of arguments in the cell.
func (sc *StackCell) ArgsCount() int {
	return len(sc.Args)
}

// HasArgs checks if the cell has enough arguments.
func (sc *StackCell) HasArgs() bool {
	required := Ops[sc.Op.Op].Args
	if sc.Op.Op == Call {
		// Amount of args needed for function call itself
		if len(sc.Args) == 0 {
			return false
		}
		identifier := sc.Args[0].String()
		if !Functs.Has(identifier) {
			return true
		}
		required += Functs[identifier].Args
	}
	return sc.ArgsCount() == required
}

func newVMStackLoop(vm *VM) VMLoop {
	return &VMStackLoop{
		vm: vm,
	}
}

func (vm *VMStackLoop) fetch() (cell Cell, err error) {
	cell, err = vm.vm.Pop()
	return
}

func (vm *VMStackLoop) Empty() bool {
	return vm.stack.Empty()
}

func (vm *VMStackLoop) Push(c *StackCell) {
	vm.stack = vm.stack.Push(c)
}

func (vm *VMStackLoop) Peek() (cell *StackCell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	cell = vm.stack.Peek()
	return
}

func (vm *VMStackLoop) Pop() (cell *StackCell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.stack, cell = vm.stack.Pop()
	return
}

func (vm *VMStackLoop) execute(cell *StackCell) (err error) {
	result, err := Ops[cell.Op.Op].Do(vm.vm, cell.Op, cell.Args...)
	if result != nil {
		vm.vm.Push(*result)
	}
	return
}

// rewindVMState puts the waiting cells back to the vm stack
func (vm *VMStackLoop) rewindVMState() (op OpCode) {
	waitingCell, _ := vm.Pop()
	op = waitingCell.Op.Op
	for _, arg := range waitingCell.Args {
		vm.vm.Push(arg)
	}
	for ; !vm.Empty(); waitingCell, _ = vm.Pop() {
		vm.vm.Push(waitingCell.Op)
		for _, arg := range waitingCell.Args {
			vm.vm.Push(arg)
		}
	}
	return
}

func (vm *VMStackLoop) Step() (err error) {
	cell, err := vm.fetch()
	if err != nil {
		if !vm.Empty() {
			// we have waiting cells, rewind them amd return error
			err = fmt.Errorf("%w: %s", ErrNotEnoughArguments, vm.rewindVMState())
		}
		return
	}
	switch {
	case cell.Op.IsOperation():
		sc := &StackCell{Op: cell}
		if sc.HasArgs() {
			err = vm.execute(sc)
			return
		}
		vm.Push(sc)
	default:
		op, popErr := vm.Pop()
		if popErr != nil {
			vm.vm.Push(cell)
			err = ErrHalt
			return
		}
		op.AddArg(cell)
		if op.HasArgs() {
			err = vm.execute(op)
		} else {
			vm.Push(op)
		}
	}
	return
}
