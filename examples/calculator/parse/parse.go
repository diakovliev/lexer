package parse

import (
	"bytes"
	"math"
	"strconv"
)

// parse float in arbitrary base
// Exponents chart:
// Pos:       ...  |0   1   2   3   4    5    6   | ...
// Digits:    ...  |N   N   N   .   N    N    N   | ...
// Exponents: ...  |b^2 b^1 b^0     b^-1 b^-2 b^-3| ...
func parseFloat(buffer []byte, base int) (result float64, err error) {
	// I hope we have enough precision)
	dotPos := bytes.IndexRune(buffer, RadixPoint)
	maxExponent := dotPos - 1
	startBase := math.Pow(float64(base), float64(maxExponent))
	for i := 0; i < len(buffer); i++ {
		if i == dotPos {
			continue
		}
		w, wErr := digitWeight(rune(buffer[i]))
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

func ParseNumber(buffer []byte) (any, error) {
	isNegative := bytes.HasPrefix(buffer, []byte("-"))
	if isNegative {
		buffer = buffer[1:]
	} else if bytes.HasPrefix(buffer, []byte("+")) {
		buffer = buffer[1:]
	}
	base := 10
	for prefix, pBase := range bases {
		if bytes.Contains(buffer, []byte(prefix)) {
			base = pBase
			buffer = buffer[len(prefix):]
			break
		}
	}
	if !bytes.ContainsFunc(buffer, IsRadixPoint) {
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
