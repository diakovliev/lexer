package parse

import (
	"bytes"
	"math"
	"strconv"

	"github.com/diakovliev/lexer/examples/calculator/number"
)

// parse float in arbitrary base
// Exponents chart:
// Pos:       ...  |0   1   2   3   4    5    6   | ...
// Digits:    ...  |N   N   N   .   N    N    N   | ...
// Exponents: ...  |b^2 b^1 b^0     b^-1 b^-2 b^-3| ...
func parseFloat(buffer []byte, base int) (result float64, err error) {
	// I hope we have enough precision)
	dotPos := bytes.IndexRune(buffer, number.Radix)
	maxExponent := dotPos - 1
	startBase := math.Pow(float64(base), float64(maxExponent))
	for i := 0; i < len(buffer); i++ {
		if i == dotPos {
			continue
		}
		w, wErr := number.Weight(rune(buffer[i]))
		if wErr != nil {
			err = wErr
			return
		}
		delta := w * startBase
		if math.IsNaN(delta) {
			// no sense to continue
			break
		}
		result += delta
		startBase /= float64(base)
	}
	return
}

// ParseNumber parses a number from the given buffer and returns it as an any value.
func ParseNumber(buffer []byte) (any, error) {
	isNegative := bytes.HasPrefix(buffer, []byte("-"))
	if isNegative {
		buffer = buffer[1:]
	} else if bytes.HasPrefix(buffer, []byte("+")) {
		buffer = buffer[1:]
	}
	base, buffer := number.DetectBase(buffer)
	if !bytes.ContainsFunc(buffer, number.IsRadix) {
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
