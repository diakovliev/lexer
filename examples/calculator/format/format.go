package format

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/diakovliev/lexer/examples/calculator/parse"
)

var (
	// ErrInvalidBase is returned when the given base is not supported by this formatter.
	ErrInvalidBase = fmt.Errorf("invalid base")

	basePrefixes = map[int]string{
		2:  parse.BinPrefixes[0],
		8:  parse.OctPrefixes[0],
		10: "",
		16: parse.HexPrefixes[0],
	}
)

func sign(f float64) float64 {
	isNegative := math.Signbit(f)
	if isNegative {
		return -1
	}
	return 1
}

// FormatNumber formats the given number in the specified base and precision.
func FormatNumber(f float64, prec uint, base int) (ret string, err error) {
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
		builder.WriteRune(parse.RadixPoint)
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
