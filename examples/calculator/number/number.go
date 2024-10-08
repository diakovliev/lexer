package number

import (
	"bytes"
	"fmt"
)

const (
	// Radix is the radix point rune.
	Radix = rune('.')

	// Bin is the binary base.
	Bin = 2
	// Oct is the octal base.
	Oct = 8
	// Dec is the decimal base.
	Dec = 10
	// Hex is the hexadecimal base.
	Hex = 16
)

type digit struct {
	runes  []rune
	weight float64
	bases  []int
}

var (
	// ErrInvalidDigit is returned when a digit is invalid.
	ErrInvalidDigit = fmt.Errorf("invalid digit")

	// DecPrefixes is a list of decimal prefixes.
	DecPrefixes = []string{}
	// BinPrefixes is a list of binary prefixes.
	BinPrefixes = []string{"0b", "0B"}
	// OctPrefixes is a list of octal prefixes.
	OctPrefixes = []string{"0o", "0O"}
	// HexPrefixes is a list of hexadecimal prefixes.
	HexPrefixes = []string{"0x", "0X"}

	// bases is a map of prefixes to bases.
	bases = map[string]int{
		BinPrefixes[0]: Bin, BinPrefixes[1]: Bin,
		OctPrefixes[0]: Oct, OctPrefixes[1]: Oct,
		HexPrefixes[0]: Hex, HexPrefixes[1]: Hex,
	}

	// digits is a list of digits.
	digits = []digit{
		{runes: []rune{'0'}, weight: 0, bases: []int{Bin, Oct, Dec, Hex}},
		{runes: []rune{'1'}, weight: 1, bases: []int{Bin, Oct, Dec, Hex}},
		{runes: []rune{'2'}, weight: 2, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'3'}, weight: 3, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'4'}, weight: 4, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'5'}, weight: 5, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'6'}, weight: 6, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'7'}, weight: 7, bases: []int{Oct, Dec, Hex}},
		{runes: []rune{'8'}, weight: 8, bases: []int{Dec, Hex}},
		{runes: []rune{'9'}, weight: 9, bases: []int{Dec, Hex}},
		{runes: []rune{'A', 'a'}, weight: 10, bases: []int{Hex}},
		{runes: []rune{'B', 'b'}, weight: 11, bases: []int{Hex}},
		{runes: []rune{'C', 'c'}, weight: 12, bases: []int{Hex}},
		{runes: []rune{'D', 'd'}, weight: 13, bases: []int{Hex}},
		{runes: []rune{'E', 'e'}, weight: 14, bases: []int{Hex}},
		{runes: []rune{'F', 'f'}, weight: 15, bases: []int{Hex}},
	}
)

// IsRadix returns true if the rune is a radix point.
func IsRadix(r rune) bool {
	return r == Radix
}

// isRadixDigit returns true if the rune is a radix point.
func isBaseDigit(base int) func(r rune) bool {
	return func(r rune) bool {
		for _, digit := range digits {
			found := false
			for _, rs := range digit.runes {
				found = found || rs == r
				if found {
					break
				}
			}
			if !found {
				continue
			}
			for _, bs := range digit.bases {
				if bs == base {
					return true
				}
			}
			return false
		}
		return false
	}
}

// Weight returns the weight of a radix point.
func Weight(r rune) (float64, error) {
	for _, digit := range digits {
		found := false
		for _, rs := range digit.runes {
			if r == rs {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		return digit.weight, nil
	}
	return 0, ErrInvalidDigit
}

// IsDecDigit returns true if the rune is a decimal digit.
func IsDecDigit(r rune) bool {
	return isBaseDigit(Dec)(r)
}

// IsHexDigit returns true if the rune is a hexadecimal digit.
func IsBinDigit(r rune) bool {
	return isBaseDigit(Bin)(r)
}

// IsOctDigit returns true if the rune is an octal digit.
func IsOctDigit(r rune) bool {
	return isBaseDigit(Oct)(r)
}

// IsHexDigit returns true if the rune is a hexadecimal digit.
func IsHexDigit(r rune) bool {
	return isBaseDigit(Hex)(r)
}

// IsPlusOrMinus returns true if the rune is a plus or minus sign.
func IsPlusOrMinus(r rune) bool {
	return r == '+' || r == '-'
}

// DetectBase returns the base of a number. The returned slice is the remaining bytes after the base has been detected.
func DetectBase(buffer []byte) (base int, newBuffer []byte) {
	base = Dec
	newBuffer = buffer
	for prefix, pBase := range bases {
		if bytes.HasPrefix(buffer, []byte(prefix)) {
			base = pBase
			newBuffer = buffer[len(prefix):]
			break
		}
	}
	return
}
