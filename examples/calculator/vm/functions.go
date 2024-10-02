package vm

import (
	"fmt"
	"math"
)

type (
	Function struct {
		Args int
		Do   func(vm *VM, args ...Cell) (result *Cell, err error)
		Desc string
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

func sum(vm *VM, _ ...Cell) (result *Cell, err error) {
	var res float64
	for {
		cell, err := vm.Pop()
		if err != nil {
			break
		}
		if cell.Op != Val {
			vm.Push(cell)
			break
		}
		res += cell.AsFloat64()
	}
	result = &Cell{
		Op:    Val,
		Value: res,
	}
	return
}

func product(vm *VM, _ ...Cell) (result *Cell, err error) {
	var res float64 = 1.
	for {
		cell, err := vm.Pop()
		if err != nil {
			break
		}
		if cell.Op != Val {
			vm.Push(cell)
			break
		}
		res *= cell.AsFloat64()
	}
	result = &Cell{
		Op:    Val,
		Value: res,
	}
	return
}

func inv(_ *VM, args ...Cell) (result *Cell, err error) {
	result = &Cell{
		Op:    Val,
		Value: 1. / args[0].AsFloat64(),
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

func help(vm *VM, _ ...Cell) (result *Cell, err error) {
	fmt.Printf("Supported functions and constants:\n")
	for name, f := range Functs {
		fmt.Printf(" %s - %s\n", name, f.Desc)
	}
	err = ErrHalt
	return
}

func reset(vm *VM, _ ...Cell) (result *Cell, err error) {
	vm.Reset()
	err = ErrHalt
	return
}

func debug(vm *VM, _ ...Cell) (result *Cell, err error) {
	vm.ToggleDebug()
	if vm.IsDebug() {
		fmt.Println("debug is on")
	} else {
		fmt.Println("debug is off")
	}
	err = ErrHalt
	return
}

func state(vm *VM, _ ...Cell) (result *Cell, err error) {
	vm.PrintState("state: ")
	err = ErrHalt
	return
}

func pop(vm *VM, _ ...Cell) (result *Cell, err error) {
	vm.Pop()
	err = ErrHalt
	return
}

func set(vm *VM, args ...Cell) (result *Cell, err error) {
	call, err := vm.Pop()
	if err != nil {
		err = fmt.Errorf("set: not enough arguments on a stack")
		return
	}
	// We expect strictly CALL
	if call.Op != Call {
		vm.Push(call)
		err = fmt.Errorf("set: invalid arguments on a stack")
		return
	}
	identifier := call.Value.(string)
	if Functs.Has(identifier) {
		err = fmt.Errorf("set: %s is a reserved name, variable is not set", identifier)
		return
	}
	vm.SetVar(identifier, args[0])
	result, ok := vm.GetVar(identifier)
	if !ok {
		err = fmt.Errorf("set: %s variable is not set", identifier)
	}
	return
}

func init() {
	Functs = Functions{
		// Constants
		"Pi":  Function{Args: 0, Do: Pi, Desc: "PI"},
		"E":   Function{Args: 0, Do: E, Desc: "E"},
		"Phi": Function{Args: 0, Do: Phi, Desc: "Phi"},
		// Exponential
		"pow": Function{Args: 2, Do: pow, Desc: "x to the power of y"},
		"exp": Function{Args: 1, Do: exp, Desc: "exponential function (e^x)"},
		// Trigonometric
		"sqrt": Function{Args: 1, Do: sqrt, Desc: "square root function (√x)"},
		"sin":  Function{Args: 1, Do: sin, Desc: "sine function (sin(x))"},
		"cos":  Function{Args: 1, Do: cos, Desc: "cosine function (cos(x))"},
		"tan":  Function{Args: 1, Do: tan, Desc: "tangent function (tan(x))"},
		"asin": Function{Args: 1, Do: asin, Desc: "arcsine function (arcsin(x))"},
		"acos": Function{Args: 1, Do: acos, Desc: "arccosine function (arccos(x))"},
		"atan": Function{Args: 1, Do: atan, Desc: "arctangent function (arctan(x))"},
		// Hyperbolic
		"sinh":  Function{Args: 1, Do: sinh, Desc: "hyperbolic sine function (sinh(x))"},
		"cosh":  Function{Args: 1, Do: cosh, Desc: "hyperbolic cosine function (cosh(x))"},
		"tanh":  Function{Args: 1, Do: tanh, Desc: "hyperbolic tangent function (tanh(x))"},
		"asinh": Function{Args: 1, Do: asinh, Desc: "hyperbolic arcsine function (arcsinh(x))"},
		"acosh": Function{Args: 1, Do: acosh, Desc: "hyperbolic arccosine function (arccosh(x))"},
		"atanh": Function{Args: 1, Do: atanh, Desc: "hyperbolic arctangent function (arctanh(x))"},
		// Log
		"log":   Function{Args: 1, Do: log, Desc: "natural logarithm function (ln(x))"},
		"log10": Function{Args: 1, Do: log10, Desc: "logarithm base 10 function (log10(x))"},
		"log2":  Function{Args: 1, Do: log2, Desc: "logarithm base 2 function (log2(x))"},
		// Miscellaneous
		"sum":     Function{Args: 0, Do: sum, Desc: "eliminate all values from the top of the stack until the first non value into theirs sum"},
		"product": Function{Args: 0, Do: product, Desc: "eliminate all values from the top of the stack until the first non value into theirs product"},
		"inv":     Function{Args: 1, Do: inv, Desc: "inversion (1/x)"},
		"abs":     Function{Args: 1, Do: abs, Desc: "absolute value function (|x|)"},
		"min":     Function{Args: 2, Do: min, Desc: "minimum function (min(a, b))"},
		"max":     Function{Args: 2, Do: max, Desc: "maximum function (max(a, b))"},
		"floor":   Function{Args: 1, Do: floor, Desc: "floor function (∌x)"},
		"ceil":    Function{Args: 1, Do: ceil, Desc: "ceiling function (^x)"},
		"round":   Function{Args: 1, Do: round, Desc: "rounding function (round(x))"},
		// System
		"help":  Function{Args: 0, Do: help, Desc: "display this help message"},
		"reset": Function{Args: 0, Do: reset, Desc: "reset the calculator to its initial state"},
		"debug": Function{Args: 0, Do: debug, Desc: "toggle debugging mode on or off"},
		"state": Function{Args: 0, Do: state, Desc: "display the current calculator state"},
		"pop":   Function{Args: 0, Do: pop, Desc: "remove the top element from the stack"},
		"set":   Function{Args: 1, Do: set, Desc: "set a variable to the given value"},
	}
}

var Functs Functions
