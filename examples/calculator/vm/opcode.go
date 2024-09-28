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
)

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
	default:
		panic("unknown opcode")
	}
}
