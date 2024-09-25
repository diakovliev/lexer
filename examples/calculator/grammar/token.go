package grammar

// Token is a lexer token.
type Token uint

const (
	// Number is a number token.
	Number Token = iota
	// Plus is a plus token.
	Plus
	// Minus is a minus token.
	Minus
	// Mul is a multiplication token.
	Mul
	// Div is a division token.
	Div
	// Bra is an opening bracket token.
	Bra
	// Ket is a closing bracket token.
	Ket
)

// String returns the string representation of a token.
func (t Token) String() string {
	switch t {
	case Number:
		return "Number"
	case Plus:
		return "Plus"
	case Minus:
		return "Minus"
	case Mul:
		return "Mul"
	case Div:
		return "Div"
	case Bra:
		return "Bra"
	case Ket:
		return "Ket"
	default:
		panic("unreachable")
	}
}
