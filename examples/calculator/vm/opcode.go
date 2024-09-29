package vm

type (
	// OpCode is a operation code
	OpCode uint
)

const (
	// Val is a value it marks the token as a value.
	Val OpCode = iota
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
	// Ident is an identifier.
	Ident
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
	case Ident:
		return "IDENT"
	default:
		panic("unknown opcode")
	}
}
