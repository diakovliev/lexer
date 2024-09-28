package vm

import (
	"errors"
	"fmt"
	"slices"

	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// VM is a simple stack virtual machine
	VM struct {
		stack stack.Stack[Cell]
	}
)

// New creates new virtual machine with given code.
func New() (vm *VM) {
	vm = &VM{}
	return
}

func (vm *VM) PushCode(code []Cell) *VM {
	for _, token := range code {
		vm.stack = vm.stack.Push(token)
	}
	return vm
}

func (vm *VM) PrintCode() *VM {
	codeCopy := make([]Cell, len(vm.stack))
	copy(codeCopy, vm.stack)
	slices.Reverse(codeCopy)
	for i, c := range codeCopy {
		if i == 0 {
			fmt.Printf(" * %04d %+v\n", i, c)
		} else {
			fmt.Printf("   %04d %+v\n", i, c)
		}
	}
	return vm
}

// Push pushes new token to the stack of virtual machine.
func (vm *VM) Push(t Cell) {
	vm.stack = vm.stack.Push(t)
}

// Pop pops token from the stack of virtual machine and returns it.
// If there is no tokens in the stack, then error will be returned.
func (vm *VM) Pop() (value Cell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.stack, value = vm.stack.Pop()
	return
}

// Peek peeks token from the top of the stack not poping it.
func (vm *VM) Peek() (value Cell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	value = vm.stack.Peek()
	return
}

func (vm *VM) fetchCommand() (cmd Cell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.stack, cmd = vm.stack.Pop()
	if !Ops.Has(cmd.Op) {
		// Return cell to the stack
		vm.Push(cmd)
		// Non operation -> halt
		err = ErrHalt
	}
	return
}

func (vm *VM) fetch() (token Cell, err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.stack, token = vm.stack.Pop()
	if Ops.Has(token.Op) {
		vm.stack = vm.stack.Push(token)
		if err = vm.step(); err != nil && !errors.Is(err, ErrHalt) {
			return
		}
		vm.stack, token = vm.stack.Pop()
		err = nil
	}
	return
}

func (vm *VM) execute(cmd Cell) (err error) {
	operation := Ops[cmd.Op]
	var arguments []Cell
	for i := 0; i < operation.Args; i++ {
		var argument Cell
		if argument, err = vm.fetch(); err != nil {
			// can't execute operation, not enough arguments
			// push fetched argiments back and return error
			if len(arguments) > 0 {
				for _, arg := range arguments {
					vm.stack = vm.stack.Push(arg)
				}
			}
			err = fmt.Errorf("%w: %s", ErrNotEnoughArguments, cmd.Op)
			return
		}
		arguments = append(arguments, argument)
	}

	result, err := Ops[cmd.Op].Do(arguments...)
	if vm.stack.Empty() {
		err = ErrHalt
	}
	// push result
	vm.stack = vm.stack.Push(result)
	return
}

func (vm *VM) step() (err error) {
	if vm.stack.Empty() {
		err = ErrStackEmpty
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
	top, err := vm.Peek()
	if err != nil {
		return
	}
	if !Ops.Has(top.Op) {
		// nothing to do, halt immediately
		err = ErrHalt
		return
	}
	for err == nil {
		err = vm.step()
	}
	return
}
