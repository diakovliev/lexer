package vm

import (
	"fmt"
	"math"
)

type (
	Function struct {
		Args int
		Do   func(vm *VM, args ...Cell) (result *Cell, err error)
	}

	Functions map[string]Function
)

func (f Functions) Has(name string) bool {
	_, ok := f[name]
	return ok
}

func (f Functions) Get(name string) Function {
	return f[name]
}

// ================ Constants ===========================
func Pi(_ *VM, _ ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Pi,
	}
	return
}

func E(_ *VM, _ ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.E,
	}
	return
}

func Phi(_ *VM, _ ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Phi,
	}
	return
}

// ================ Functions ===========================
func pow(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Pow(args[0].AsFloat64(), args[1].AsFloat64()),
	}
	return
}

func exp(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Exp(args[0].AsFloat64()),
	}
	return
}

func sqrt(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Sqrt(args[0].AsFloat64()),
	}
	return
}

func sin(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Sin(args[0].AsFloat64()),
	}
	return
}

func cos(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Cos(args[0].AsFloat64()),
	}
	return
}

func tan(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Tan(args[0].AsFloat64()),
	}
	return
}

func asin(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Asin(args[0].AsFloat64()),
	}
	return
}

func acos(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Acos(args[0].AsFloat64()),
	}
	return
}

func atan(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Atan(args[0].AsFloat64()),
	}
	return
}

func sinh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Sinh(args[0].AsFloat64()),
	}
	return
}

func cosh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Cosh(args[0].AsFloat64()),
	}
	return
}

func tanh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Tanh(args[0].AsFloat64()),
	}
	return
}

func asinh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Asinh(args[0].AsFloat64()),
	}
	return
}

func acosh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Acosh(args[0].AsFloat64()),
	}
	return
}

func atanh(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Atanh(args[0].AsFloat64()),
	}
	return
}

func log(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Log(args[0].AsFloat64()),
	}
	return
}

func log10(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Log10(args[0].AsFloat64()),
	}
	return
}

func log2(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Log2(args[0].AsFloat64()),
	}
	return
}

func abs(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Abs(args[0].AsFloat64()),
	}
	return
}

func min(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Min(args[0].AsFloat64(), args[1].AsFloat64()),
	}
	return
}

func max(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Max(args[0].AsFloat64(), args[1].AsFloat64()),
	}
	return
}

func floor(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Floor(args[0].AsFloat64()),
	}
	return
}

func ceil(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Ceil(args[0].AsFloat64()),
	}
	return
}

func round(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: math.Round(args[0].AsFloat64()),
	}
	return
}

func reset(vm *VM, _ ...Cell) (result *Cell, err error) {
	vm.Reset()
	err = ErrHalt
	return
}

func set(vm *VM, args ...Cell) (result *Cell, err error) {
	call, err := vm.Pop()
	if err != nil {
		err = fmt.Errorf("set: not enough arguments on a stack")
		return
	}
	identifier, err := vm.Pop()
	if err != nil {
		err = fmt.Errorf("set: not enough arguments on a stack")
		return
	}
	// We expect strictly CALL <identifier> sequence on stack
	if call.Op != Call || identifier.Op != Ident {
		err = fmt.Errorf("set: invalid arguments on a stack")
		return
	}
	if Functs.Has(identifier.String()) {
		err = fmt.Errorf("set: %s is a reserved name, varible is not set", identifier.String())
		return
	}
	vm.SetVar(identifier, args[0])
	result, ok := vm.GetVar(identifier)
	if !ok {
		err = fmt.Errorf("set: %s varible is not set", identifier.String())
	}
	return
}

func init() {
	Functs = Functions{
		// Constants
		"Pi":  Function{Args: 0, Do: Pi},
		"E":   Function{Args: 0, Do: E},
		"Phi": Function{Args: 0, Do: Phi},
		// Exponential
		"pow": Function{Args: 2, Do: pow},
		"exp": Function{Args: 1, Do: exp},
		// Trigonometric
		"sqrt": Function{Args: 1, Do: sqrt},
		"sin":  Function{Args: 1, Do: sin},
		"cos":  Function{Args: 1, Do: cos},
		"tan":  Function{Args: 1, Do: tan},
		"asin": Function{Args: 1, Do: asin},
		"acos": Function{Args: 1, Do: acos},
		"atan": Function{Args: 1, Do: atan},
		// Hyperbolic
		"sinh":  Function{Args: 1, Do: sinh},
		"cosh":  Function{Args: 1, Do: cosh},
		"tanh":  Function{Args: 1, Do: tanh},
		"asinh": Function{Args: 1, Do: asinh},
		"acosh": Function{Args: 1, Do: acosh},
		"atanh": Function{Args: 1, Do: atanh},
		// Log
		"log":   Function{Args: 1, Do: log},
		"log10": Function{Args: 1, Do: log10},
		"log2":  Function{Args: 1, Do: log2},
		// Miscellaneous
		"abs":   Function{Args: 1, Do: abs},
		"min":   Function{Args: 2, Do: min},
		"max":   Function{Args: 2, Do: max},
		"floor": Function{Args: 1, Do: floor},
		"ceil":  Function{Args: 1, Do: ceil},
		"round": Function{Args: 1, Do: round},
		// System
		"reset": Function{Args: 0, Do: reset},
		"set":   Function{Args: 1, Do: set},
	}
}

var Functs Functions
