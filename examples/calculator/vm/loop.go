package vm

import (
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// LoopStackCell
	LoopStackCell struct {
		Op   Cell
		Args []Cell
	}

	Loop struct {
		vm        *VM
		stack     stack.Stack[*LoopStackCell]
		backStack stack.Stack[Cell]
	}
)

// AddArg adds new argument to the cell.
func (sc *LoopStackCell) AddArg(c Cell) {
	sc.Args = append(sc.Args, c)
}

// ArgsCount returns the number of arguments in the cell.
func (sc *LoopStackCell) ArgsCount() int {
	return len(sc.Args)
}

// HasArgs checks if the cell has enough arguments.
func (sc *LoopStackCell) HasArgs() bool {
	required := Ops[sc.Op.Op].Args
	if sc.Op.Op == Call {
		identifier := sc.Op.Value.(string)
		if !Functs.Has(identifier) {
			return true
		}
		required += Functs[identifier].Args
	}
	return sc.ArgsCount() == required
}

func newVMStackLoop(vm *VM) *Loop {
	return &Loop{
		vm: vm,
	}
}

func (loop *Loop) PrintState(pfx string) {
	if len(loop.stack) > 0 {
		fmt.Fprintf(loop.vm.output, "%sOps stack:\n", pfx)
		for i := 0; i < len(loop.stack); i++ {
			sc := loop.stack[i]
			fmt.Fprintf(loop.vm.output, "%s    %d %s\t%s\n", pfx, i, sc.Op.String(), loop.vm.formatCellValue(sc.Op))
		}
	}
}

func (loop *Loop) fetch() (cell Cell, err error) {
	cell, err = loop.vm.Pop()
	if errors.Is(err, ErrStackEmpty) {
		err = ErrHalt
	}
	return
}

func (loop *Loop) Empty() bool {
	return loop.stack.Empty()
}

func (loop *Loop) Push(c *LoopStackCell) {
	loop.stack = loop.stack.Push(c)
}

func (loop *Loop) Peek() (cell *LoopStackCell, err error) {
	if loop.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	cell = loop.stack.Peek()
	return
}

func (loop *Loop) Pop() (cell *LoopStackCell, err error) {
	if loop.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	loop.stack, cell = loop.stack.Pop()
	return
}

func (loop *Loop) execute(cell *LoopStackCell) (err error) {
	result, err := Ops[cell.Op.Op].Do(loop.vm, cell.Op, cell.Args...)
	if err != nil && !errors.Is(err, ErrHalt) {
		loop.Push(cell)
		loop.rewindVMState()
	}
	if result != nil {
		loop.vm.Push(*result)
	}
	return
}

// rewindVMState puts the waiting cells back to the vm stack
func (loop *Loop) rewindVMState() (op OpCode) {
	waitingCell, _ := loop.Pop()
	op = waitingCell.Op.Op
	for i := len(waitingCell.Args) - 1; i >= 0; i-- {
		loop.vm.Push(waitingCell.Args[i])
	}
	for ; !loop.Empty(); waitingCell, _ = loop.Pop() {
		loop.vm.Push(waitingCell.Op)
		for i := len(waitingCell.Args) - 1; i >= 0; i-- {
			loop.vm.Push(waitingCell.Args[i])
		}
	}
	return
}

func (loop *Loop) Finalize() {
	var cell Cell
	for !loop.backStack.Empty() {
		loop.backStack, cell = loop.backStack.Pop()
		loop.vm.Push(cell)
	}
}

func (loop *Loop) Step() (err error) {
	cell, err := loop.fetch()
	if err != nil {
		if !loop.Empty() {
			// we have waiting cells, rewind them amd return error
			err = fmt.Errorf("%w: %s", ErrNotEnoughArguments, loop.rewindVMState())
		}
		return
	}
	switch {
	case cell.Op.IsOperation():
		sc := &LoopStackCell{Op: cell}
		if sc.HasArgs() {
			err = loop.execute(sc)
			return
		}
		loop.Push(sc)
	default:
		op, popErr := loop.Pop()
		if popErr != nil {
			loop.backStack = loop.backStack.Push(cell)
			return
		}
		op.AddArg(cell)
		if op.HasArgs() {
			err = loop.execute(op)
		} else {
			loop.Push(op)
		}
	}
	return
}
