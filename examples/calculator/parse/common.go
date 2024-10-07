package parse

import "fmt"

const (
	// FIXME: Detect the current locale radix point and use it instead of '.'.
	RadixPoint = rune('.')
)

// IsRadixPoint returns true if the rune is a radix point.
func IsRadixPoint(r rune) bool {
	return r == RadixPoint
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

	bases = map[string]int{
		BinPrefixes[0]: 2, BinPrefixes[1]: 2,
		OctPrefixes[0]: 8, OctPrefixes[1]: 8,
		HexPrefixes[0]: 16, HexPrefixes[1]: 16,
	}
)

type digit struct {
	runes  []rune
	weight float64
	bases  []int
}

var digits = []digit{
	{runes: []rune{'0'}, weight: 0, bases: []int{2, 8, 10, 16}},
	{runes: []rune{'1'}, weight: 1, bases: []int{2, 8, 10, 16}},
	{runes: []rune{'2'}, weight: 2, bases: []int{8, 10, 16}},
	{runes: []rune{'3'}, weight: 3, bases: []int{8, 10, 16}},
	{runes: []rune{'4'}, weight: 4, bases: []int{8, 10, 16}},
	{runes: []rune{'5'}, weight: 5, bases: []int{8, 10, 16}},
	{runes: []rune{'6'}, weight: 6, bases: []int{8, 10, 16}},
	{runes: []rune{'7'}, weight: 7, bases: []int{8, 10, 16}},
	{runes: []rune{'8'}, weight: 8, bases: []int{10, 16}},
	{runes: []rune{'9'}, weight: 9, bases: []int{10, 16}},
	{runes: []rune{'A', 'a'}, weight: 10, bases: []int{16}},
	{runes: []rune{'B', 'b'}, weight: 11, bases: []int{16}},
	{runes: []rune{'C', 'c'}, weight: 12, bases: []int{16}},
	{runes: []rune{'D', 'd'}, weight: 13, bases: []int{16}},
	{runes: []rune{'E', 'e'}, weight: 14, bases: []int{16}},
	{runes: []rune{'F', 'f'}, weight: 16, bases: []int{16}},
}

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

func digitWeight(r rune) (float64, error) {
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

func IsDecDigit(r rune) bool {
	return isBaseDigit(10)(r)
}

func IsBinDigit(r rune) bool {
	return isBaseDigit(2)(r)
}

func IsOctDigit(r rune) bool {
	return isBaseDigit(8)(r)
}

func IsHexDigit(r rune) bool {
	return isBaseDigit(16)(r)
}

func IsPlusOrMinus(r rune) bool {
	return r == '+' || r == '-'
}
