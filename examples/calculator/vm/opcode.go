package vm

type (
	// OpCode is a operation code
	OpCode uint
)

const (
	// Comma is a comma. Used only on parsing.
	Comma OpCode = iota
	// Bra is an opening bracket. Used only on parsing.
	Bra
	// Ket is a closing bracket. Used only on parsing.
	Ket
	// Val is a value it marks the cell as a value.
	Val
	// Add is a add operation.
	Add
	// Sub is a minus operation.
	Sub
	// Mul is a multiplication operation.
	Mul
	// Div is a division operation.
	Div
	// Call is a call operation.
	Call
)

// IsOperation checks if the opcode is an operation.
func (op OpCode) IsOperation() bool {
	return Ops.Has(op)
}

// Args returns the number of required arguments.
func (op OpCode) Args() int {
	return Ops[op].Args
}

// String returns a string representation of the opcode.
func (op OpCode) String() string {
	switch op {
	case Comma:
		return ","
	case Bra:
		return "("
	case Ket:
		return ")"
	case Val:
		return "VAL"
	case Add:
		return "ADD"
	case Sub:
		return "SUB"
	case Mul:
		return "MUL"
	case Div:
		return "DIV"
	case Call:
		return "CALL"
	default:
		panic("unknown opcode")
	}
}
