package vm

import (
	"fmt"
	"slices"

	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// VM is a simple stack virtual machine
	VM struct {
		code      stack.Stack[Cell]
		vars      map[string]Cell
		loop      *Loop
		debug     bool
		outputFmt string
	}
)

func newVM(loop func(*VM) *Loop) (vm *VM) {
	vm = &VM{
		vars:      make(map[string]Cell),
		outputFmt: "%s%s[%d]:\t%v\n",
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
	vm.code = stack.Stack[Cell]{}
	vm.vars = make(map[string]Cell)
	return
}

func (vm *VM) PushCode(code []Cell) *VM {
	for _, cell := range code {
		vm.code = vm.code.Push(cell)
	}
	return vm
}

func (vm *VM) ToggleDebug() *VM {
	vm.debug = !vm.debug
	return vm
}

func (vm *VM) IsDebug() bool {
	return vm.debug
}

func (vm *VM) PrintState(pfx string) *VM {
	vm.loop.PrintState(pfx)
	codeCopy := make([]Cell, len(vm.code))
	copy(codeCopy, vm.code)
	slices.Reverse(codeCopy)
	if len(vm.vars) > 0 {
		fmt.Printf("%sVariables:\n", pfx)
		for k, v := range vm.vars {
			fmt.Printf("%s    %s = %+v\n", pfx, k, v)
		}
	}
	if len(codeCopy) > 0 {
		fmt.Printf("%sStack:\n", pfx)
		for i, c := range codeCopy {
			if i == 0 {
				fmt.Printf("%s  * %04d %+v\n", pfx, i, c)
			} else {
				fmt.Printf("%s    %04d %+v\n", pfx, i, c)
			}
		}
	}
	return vm
}

// Empty checks if the stack is empty.
func (vm VM) Empty() bool {
	return vm.code.Empty()
}

// Push pushes new token to the stack of virtual machine.
func (vm *VM) Push(c Cell) {
	vm.code = vm.code.Push(c)
}

// Pop pops token from the stack of virtual machine and returns it.
// If there is no tokens in the stack, then error will be returned.
func (vm *VM) Pop() (cell Cell, err error) {
	if vm.code.Empty() {
		err = ErrStackEmpty
		return
	}
	vm.code, cell = vm.code.Pop()
	return
}

// Peek peeks token from the top of the stack not poping it.
func (vm VM) Peek() (cell Cell, err error) {
	if vm.code.Empty() {
		err = ErrStackEmpty
		return
	}
	cell = vm.code.Peek()
	return
}

func (vm *VM) SetVar(identifier string, value Cell) {
	vm.vars[identifier] = value
}

func (vm *VM) CreateVar(identifier string) {
	cell := Cell{Op: Val, Value: int64(0)}
	vm.vars[identifier] = cell
}

func (vm *VM) GetVar(identifier string) (result *Cell, ok bool) {
	var value Cell
	if value, ok = vm.vars[identifier]; ok {
		result = &value
	}
	return
}

// Call calls an external function or resolves constant or variable
func (vm *VM) Call(op Cell, args ...Cell) (result *Cell, err error) {
	identifier := op.Value.(string)
	if !Functs.Has(identifier) {
		// try to resolve variable
		var ok bool
		result, ok = vm.GetVar(identifier)
		if !ok {
			err = fmt.Errorf("%w: %s", ErrUnknownIdentifier, identifier)
		}
		return
	}
	// invoke function
	result, err = Functs.Get(identifier).Do(vm, args...)
	return
}

func (vm *VM) Output(prefix, out string) {
	if len(vm.code) == 0 {
		fmt.Printf(vm.outputFmt, prefix, out, 0, " <none>")
		return
	}
	idx := 0
	for i := len(vm.code) - 1; i >= 0; i-- {
		fmt.Printf(vm.outputFmt, prefix, out, idx, vm.code[i])
		idx++
	}
}

// Run the VM, return ErrVMHalt when finished.
func (vm *VM) Run() (err error) {
	if vm.debug {
		fmt.Printf("debug: STATE BEFORE ->\n")
		vm.PrintState("debug: ")
		fmt.Printf("debug: <- STATE BEFORE\n")
	}
	defer func() {
		if vm.debug {
			fmt.Printf("debug: STATE AFTER ->\n")
			vm.PrintState("debug: ")
			fmt.Printf("debug: <- STATE AFTER\n")
		}
	}()
	for err == nil {
		err = vm.loop.Step()
	}
	vm.loop.Finalize()
	return
}
