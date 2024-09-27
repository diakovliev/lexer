package lexer_test

// Token is a lexer token.
type Token uint

const (
	// BinNumber is a binary signed or unsigned number token.
	BinNumber Token = iota
	// OctNumber is a octal signed or unsigned number token.
	OctNumber
	// DecNumber is a decimal signed or unsigned number token.
	DecNumber
	// HexNumber is a hexadecimal signed or unsigned number token.
	HexNumber
	// BinFraction is a binary signed or unsigned fraction token.
	BinFraction
	// OctFraction is a octal signed or unsigned fraction token.
	OctFraction
	// DecFraction is a decimal signed or unsigned fraction token.
	DecFraction
	// HexFraction is a hexadecimal signed or unsigned fraction token.
	HexFraction
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
	// Comma is a comma token.
	Comma
	// Identifier is an identifier token.
	Identifier
	// String is a string token.
	String
)

// String returns the string representation of a token.
func (t Token) String() string {
	switch t {
	case BinNumber:
		return "BinNumber"
	case OctNumber:
		return "OctNumber"
	case DecNumber:
		return "DecNumber"
	case HexNumber:
		return "HexNumber"
	case BinFraction:
		return "BinFraction"
	case OctFraction:
		return "OctFraction"
	case DecFraction:
		return "DecFraction"
	case HexFraction:
		return "HexFraction"
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
	case Comma:
		return "Comma"
	case Identifier:
		return "Identifier"
	case String:
		return "String"
	default:
		panic("unreachable")
	}
}
