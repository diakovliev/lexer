package vm

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/diakovliev/lexer/examples/calculator/number/format"
	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// VM is a simple stack virtual machine
	VM struct {
		code   stack.Stack[Cell]
		vars   map[string]Cell
		loop   *Loop
		debug  bool
		outFmt string
		output io.Writer
	}
)

const (
	outModeVar     = "out_mode"
	outModeHex     = 1 << 1
	outModeOct     = 1 << 2
	outModeBin     = 1 << 3
	outModeDecOnly = 0
	outModeAll     = outModeHex | outModeOct | outModeBin

	outPrecisionVar     = "out_precision"
	outDefaultPrecision = 8
)

var (
	settingsVars = map[string]Cell{
		outModeVar:      {Op: Val, Value: int64(outModeDecOnly)},
		outPrecisionVar: {Op: Val, Value: int64(outDefaultPrecision)},
	}
)

func newVM(loop func(*VM) *Loop, opts ...Option) (vm *VM) {
	vm = &VM{
		vars:   map[string]Cell{},
		outFmt: "%s%s[%d]:\t%s\n",
		output: os.Stdout,
	}
	for _, opt := range opts {
		opt(vm)
	}
	for k, v := range settingsVars {
		vm.vars[k] = v
	}
	vm.loop = loop(vm)
	return
}

// New creates new virtual machine with given code.
func New(opts ...Option) (vm *VM) {
	vm = newVM(newVMStackLoop, opts...)
	return
}

// Reset resets virtual machine to initial state, except settings vars.
func (vm *VM) Reset() (err error) {
	vm.code = stack.Stack[Cell]{}
	vm.vars = map[string]Cell{}
	for k, v := range settingsVars {
		vm.vars[k] = v
	}
	return
}

// PushCode pushes code to the virtual machine stack.
func (vm *VM) PushCode(code []Cell) *VM {
	for _, cell := range code {
		vm.code = vm.code.Push(cell)
	}
	return vm
}

// ToggleDebug enables or disables debug mode.
func (vm *VM) ToggleDebug() *VM {
	vm.debug = !vm.debug
	return vm
}

// IsDebug returns true if debug mode is enabled.
func (vm *VM) IsDebug() bool {
	return vm.debug
}

// PrintState prints current state of the virtual machine to output.
func (vm *VM) PrintState(pfx string) *VM {
	vm.loop.PrintState(pfx)
	codeCopy := make([]Cell, len(vm.code))
	copy(codeCopy, vm.code)
	slices.Reverse(codeCopy)
	if len(vm.vars) > 0 {
		fmt.Fprintf(vm.output, "%sVariables:\n", pfx)
		for k, v := range vm.vars {
			fmt.Fprintf(vm.output, "%s    %s =\t%s\n", pfx, k, vm.formatCellValue(v))
		}
	}
	if len(codeCopy) > 0 {
		fmt.Fprintf(vm.output, "%sStack:\n", pfx)
		for i, c := range codeCopy {
			fmtStr := "%s    %04d\t%s\n"
			if i == 0 {
				fmtStr = "%s  * %04d\t%s\n"
			}
			fmt.Fprintf(vm.output, fmtStr, pfx, i, vm.formatCellValue(c))
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

// SetVar sets variable to the given value. It also updates settingsVars map.
func (vm *VM) SetVar(identifier string, value Cell) {
	vm.vars[identifier] = value
	if _, ok := settingsVars[identifier]; ok {
		settingsVars[identifier] = value
	}
}

// CreateVar creates new variable with given identifier and sets it to zero.
func (vm *VM) CreateVar(identifier string) {
	cell := Cell{Op: Val, Value: int64(0)}
	vm.vars[identifier] = cell
}

// GetVar returns cell with given identifier and true if it exists in the vars map. Otherwise false.
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

func (vm VM) getOutBases() (bases []int) {
	bases = append(bases, 10)
	cell := vm.vars[outModeVar]
	mode := cell.AsInt64()
	if mode&outModeHex != 0 {
		bases = append(bases, 16)
	}
	if mode&outModeOct != 0 {
		bases = append(bases, 8)
	}
	if mode&outModeBin != 0 {
		bases = append(bases, 2)
	}
	return
}

func (vm VM) getOutPrecision() uint {
	cell := vm.vars[outPrecisionVar]
	return uint(cell.AsInt64())
}

func (vm VM) formatCellValue(cell Cell) (ret string) {
	builder := strings.Builder{}
	if !cell.IsNumber() {
		ret = cell.String()
		return
	}
	precision := vm.getOutPrecision()
	outBases := vm.getOutBases()
	for _, base := range outBases {
		str, err := format.FormatNumber(cell.AsFloat64(), precision, base)
		if err != nil {
			builder.WriteString("<")
			builder.WriteString(err.Error())
			builder.WriteString(">")
		} else {
			builder.WriteString(str)
		}
		builder.WriteByte('\t')
	}
	ret = strings.TrimSpace(builder.String())
	return
}

// Perform output to stdout
func (vm *VM) Output(prefix, out string) {
	if len(vm.code) == 0 {
		fmt.Fprintf(vm.output, vm.outFmt, prefix, out, 0, "<none>")
		return
	}
	idx := 0
	for i := len(vm.code) - 1; i >= 0; i-- {
		fmt.Fprintf(vm.output, vm.outFmt, prefix, out, idx, vm.formatCellValue(vm.code[i]))
		idx++
	}
}

// Run the VM, return ErrVMHalt when finished.
func (vm *VM) Run() (err error) {
	if vm.debug {
		fmt.Fprintf(vm.output, "debug: STATE BEFORE ->\n")
		vm.PrintState("debug: ")
		fmt.Fprintf(vm.output, "debug: <- STATE BEFORE\n")
	}
	defer func() {
		if vm.debug {
			fmt.Fprintf(vm.output, "debug: STATE AFTER ->\n")
			vm.PrintState("debug: ")
			fmt.Fprintf(vm.output, "debug: <- STATE AFTER\n")
		}
	}()
	for err == nil {
		err = vm.loop.Step()
	}
	vm.loop.Finalize()
	return
}
