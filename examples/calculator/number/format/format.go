package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/diakovliev/lexer/examples/calculator/number"
)

var (
	// ErrInvalidBase is returned when the given base is not supported by this formatter.
	ErrInvalidBase = fmt.Errorf("invalid base")

	// basePrefixes maps a base to its prefix. For example, for binary it returns "0b".
	basePrefixes = map[int]string{
		number.Bin: number.BinPrefixes[0],
		number.Oct: number.OctPrefixes[0],
		number.Dec: "",
		number.Hex: number.HexPrefixes[0],
	}
)

// sign returns the sign of the given number. It is either -1 or 1.
func sign(f float64) float64 {
	isNegative := math.Signbit(f)
	if isNegative {
		return -1
	}
	return 1
}

// FormatNumber formats the given number in the specified base and precision.
func FormatNumber(f float64, prec uint, base int) (ret string, err error) {
	if math.IsNaN(f) {
		ret = "NaN"
		return
	}
	if math.IsInf(f, 1) {
		ret = "Inf"
		return
	}
	if math.IsInf(f, -1) {
		ret = "-Inf"
		return
	}
	builder := strings.Builder{}
	sign := sign(f)
	if sign < 0 {
		builder.WriteRune('-')
	}
	pfx, ok := basePrefixes[base]
	if !ok {
		err = fmt.Errorf("%w: %d", ErrInvalidBase, base)
		return
	}
	if pfx != "" {
		builder.WriteString(pfx)
	}
	fAbs := math.Abs(f)
	floor := math.Floor(fAbs)
	tail := math.Abs(fAbs - floor)
	if tail == 0.0 {
		builder.WriteString(strings.ToUpper(strconv.FormatInt(int64(floor), base)))
		ret = builder.String()
		return
	}
	if prec == 0 {
		round := math.Round(fAbs + tail)
		builder.WriteString(strings.ToUpper(strconv.FormatInt(int64(round), base)))
		ret = builder.String()
		return
	} else {
		builder.WriteString(strings.ToUpper(strconv.FormatInt(int64(floor), base)))
		builder.WriteRune(number.Radix)
	}
	for i := uint(0); i < prec; i++ {
		tail *= float64(base)
		var digit int
		if i == prec-1 {
			digit = int(math.Round(tail))
		} else {
			digit = int(math.Floor(tail))
		}
		builder.WriteString(strings.ToUpper(strconv.FormatInt(int64(digit), base)))
		tail -= float64(digit)
	}
	ret = builder.String()
	return
}
