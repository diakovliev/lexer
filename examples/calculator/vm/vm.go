package vm

import (
	"fmt"
	"slices"

	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	VMLoop interface {
		Step() error
	}

	// VM is a simple stack virtual machine
	VM struct {
		state stack.Stack[Cell]
		vars  map[string]Cell
		loop  VMLoop
		debug bool
	}
)

func newVM(loop func(*VM) VMLoop) (vm *VM) {
	vm = &VM{
		vars: make(map[string]Cell),
	}
	vm.loop = loop(vm)
	return
}

// New creates new virtual machine with given code.
func New() (vm *VM) {
	vm = newVM(newVMStackLoop)
	return
}

func (vm *VM) Reset() (err error) {
	vm.state = stack.Stack[Cell]{}
	vm.vars = make(map[string]Cell)
	return
}

func (vm *VM) PushCode(code []Cell) *VM {
	for _, cell := range code {
		vm.state = vm.state.Push(cell)
	}
	return vm
}

func (vm *VM) ToggleDebug() *VM {
	vm.debug = !vm.debug
	return vm
}

func (vm *VM) PrintState() *VM {
	codeCopy := make([]Cell, len(vm.state))
	copy(codeCopy, vm.state)
	slices.Reverse(codeCopy)
	fmt.Printf("Variables:\n")
	for k, v := range vm.vars {
		fmt.Printf("   %s = %+v\n", k, v)
	}
	fmt.Printf("Stack:\n")
	for i, c := range codeCopy {
		if i == 0 {
			fmt.Printf(" * %04d %+v\n", i, c)
		} else {
			fmt.Printf("   %04d %+v\n", i, c)
		}
	}
	return vm
}

// Empty checks if the stack is empty.
func (vm VM) Empty() bool {
	return vm.state.Empty()
}

// Push pushes new token to the stack of virtual machine.
func (vm *VM) Push(c Cell) {
	vm.state = vm.state.Push(c)
}

// Pop pops token from the stack of virtual machine and returns it.
// If there is no tokens in the stack, then error will be returned.
func (vm *VM) Pop() (cell Cell, err error) {
	if vm.state.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.state, cell = vm.state.Pop()
	return
}

// Peek peeks token from the top of the stack not poping it.
func (vm VM) Peek() (cell Cell, err error) {
	if vm.state.Empty() {
		err = ErrStackEmpty
		return
	}
	cell = vm.state.Peek()
	return
}

func (vm *VM) SetVar(identifier Cell, value Cell) {
	vm.vars[identifier.String()] = value
}

func (vm *VM) CreateVar(identifier Cell) {
	cell := Cell{Op: Val, Value: int64(0)}
	vm.vars[identifier.String()] = cell
}

func (vm *VM) GetVar(identifier Cell) (result *Cell, ok bool) {
	var value Cell
	if value, ok = vm.vars[identifier.String()]; ok {
		result = &value
	}
	return
}

// Call calls an external function or resolves constant or variable
func (vm *VM) Call(op Cell, args ...Cell) (result *Cell, err error) {
	identifier := args[0]
	if !Functs.Has(identifier.String()) {
		// try to resolve variable
		var ok bool
		result, ok = vm.GetVar(identifier)
		if !ok {
			err = fmt.Errorf("%w: %s", ErrUnknownIdentifier, identifier.String())
		}
		return
	}
	// invoke function
	result, err = Functs.Get(identifier.String()).Do(vm, args[1:]...)
	return
}

// Run the VM, return ErrVMHalt when finished.
func (vm *VM) Run() (err error) {
	if vm.debug {
		fmt.Printf("STATE BEFORE ->\n")
		vm.PrintState()
		fmt.Printf("<- STATE BEFORE\n")
	}
	defer func() {
		if vm.debug {
			fmt.Printf("STATE AFTER ->\n")
			vm.PrintState()
			fmt.Printf("<- STATE AFTER\n")
		}
	}()
	cell, err := vm.Peek()
	if err != nil {
		return
	}
	if !cell.Op.IsOperation() {
		// nothing to do, halt immediately
		err = ErrHalt
		return
	}
	for err == nil {
		err = vm.loop.Step()
	}
	return
}
