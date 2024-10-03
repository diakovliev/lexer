package parse

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
)

var (
	// ErrInvalidDigit is returned when a digit is invalid.
	ErrInvalidDigit = fmt.Errorf("invalid digit")

	numberBases = map[string]int{
		grammar.BinNumberPrefixes[0]: 2,
		grammar.BinNumberPrefixes[1]: 2,
		grammar.OctNumberPrefixes[0]: 8,
		grammar.OctNumberPrefixes[1]: 8,
		grammar.HexNumberPrefixes[0]: 16,
		grammar.HexNumberPrefixes[1]: 16,
	}

	digitsValues = map[byte]float64{
		'0': 0,
		'1': 1,
		'2': 2,
		'3': 3,
		'4': 4,
		'5': 5,
		'6': 6,
		'7': 7,
		'8': 8,
		'9': 9,
		'a': 10, 'A': 10,
		'b': 11, 'B': 11,
		'c': 12, 'C': 12,
		'd': 13, 'D': 13,
		'e': 14, 'E': 14,
		'f': 15, 'F': 15,
	}
)

// parse float in arbitrary base
// Exponents chart:
// Pos:       ...  |0   1   2   3   4    5    6   | ...
// Digits:    ...  |N   N   N   .   N    N    N   | ...
// Exponents: ...  |b^2 b^1 b^0     b^-1 b^-2 b^-3| ...
func parseFloat(buffer []byte, base int) (result float64, err error) {
	// I hope we have enough precision)
	dotPos := bytes.IndexRune(buffer, grammar.Radix)
	maxExponent := dotPos - 1
	startBase := math.Pow(float64(base), float64(maxExponent))
	for i := 0; i < len(buffer); i++ {
		if i == dotPos {
			continue
		}
		dv, ok := digitsValues[buffer[i]]
		if !ok {
			err = fmt.Errorf("%w '%c'", ErrInvalidDigit, buffer[i])
			return
		}
		delta := dv * startBase
		if math.IsNaN(delta) {
			// no sense to continue
			break
		}
		result += delta
		startBase /= float64(base)
	}
	return
}

func ParseNumber(buffer []byte) (any, error) {
	isNegative := bytes.HasPrefix(buffer, []byte("-"))
	if isNegative {
		buffer = buffer[1:]
	} else if bytes.HasPrefix(buffer, []byte("+")) {
		buffer = buffer[1:]
	}
	base := 10
	for prefix, pBase := range numberBases {
		if bytes.Contains(buffer, []byte(prefix)) {
			base = pBase
			buffer = buffer[len(prefix):]
			break
		}
	}
	if !bytes.ContainsFunc(buffer, grammar.IsRadix) {
		var result int64
		// whole
		result, err := strconv.ParseInt(string(buffer), base, 64)
		if err != nil {
			return nil, err
		}
		if isNegative {
			result = -result
		}
		return result, nil
	}
	// float
	var result float64
	result, err := parseFloat(buffer, base)
	if err != nil {
		return nil, err
	}
	if isNegative {
		result = -result
	}
	return result, nil
}
