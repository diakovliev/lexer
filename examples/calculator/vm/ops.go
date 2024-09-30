package vm

type (
	// Operation is a operatation properties.
	Operation struct {
		// Args is a number of required arguments.
		Args int
		// Do is a function that performs the operation.
		Do func(vm *VM, op Cell, args ...Cell) (*Cell, error)
	}

	// Operations is a map of operators and their properties.
	Operations map[OpCode]Operation
)

// Has checks if the token is an operator.
func (o Operations) Has(op OpCode) (ok bool) {
	_, ok = o[op]
	return
}

// addOperation adds a new operation to the operations map.
func addOperation(op OpCode, operandsCount int, do func(vm *VM, op Cell, args ...Cell) (*Cell, error)) {
	operation := Operation{Args: operandsCount, Do: do}
	Ops[op] = operation
}

// Ops is a map of operators and their properties.
var Ops = Operations{}

func init() {
	addOperation(Add, 2, add)
	addOperation(Sub, 2, sub)
	addOperation(Mul, 2, mul)
	addOperation(Div, 2, div)
	addOperation(Call, 0, call)
}
