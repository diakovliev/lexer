package format

import (
	"fmt"
	"testing"

	"github.com/diakovliev/lexer/examples/calculator/parse"
	"github.com/stretchr/testify/assert"
)

func TestFormatNumber(t *testing.T) {
	type testCase struct {
		input     string
		prec      uint
		base      int
		expected  string
		wantError error
	}
	tests := []testCase{
		{
			input:    "0xABC",
			prec:     0,
			base:     16,
			expected: "0xABC",
		},
		{
			input:    "0xABC.ABC",
			prec:     0,
			base:     16,
			expected: "0xABD",
		},
		{
			input:    "0xABC.0ABC",
			prec:     3,
			base:     16,
			expected: "0xABC.0AC",
		},
		{
			input:    "-0xABC.ABC",
			prec:     0,
			base:     16,
			expected: "-0xABD",
		},
		{
			input:    "0xABC.ABC",
			prec:     1,
			base:     16,
			expected: "0xABC.B",
		},
		{
			input:    "-0xABC.ABC",
			prec:     1,
			base:     16,
			expected: "-0xABC.B",
		},
		{
			input:    "0xABC.ABC",
			prec:     2,
			base:     16,
			expected: "0xABC.AC",
		},
		{
			input:    "-0xABC.ABC",
			prec:     2,
			base:     16,
			expected: "-0xABC.AC",
		},
		{
			input:    "0xABC.ABC",
			prec:     3,
			base:     16,
			expected: "0xABC.ABC",
		},
		{
			input:    "-0xABC.ABC",
			prec:     3,
			base:     16,
			expected: "-0xABC.ABC",
		},
		{
			input:    "0xABC.ABC",
			prec:     4,
			base:     16,
			expected: "0xABC.ABC0",
		},
		{
			input:    "-0xABC.ABC",
			prec:     4,
			base:     16,
			expected: "-0xABC.ABC0",
		},
	}
	for _, tc := range tests {
		testName := fmt.Sprintf("input %s base %d prec %d expected %s", tc.input, tc.base, tc.prec, tc.expected)
		t.Run(testName, func(t *testing.T) {
			a, err := parse.ParseNumber([]byte(tc.input))
			assert.NoError(t, err)
			assert.NotNil(t, a)
			f := 0.0
			switch v := a.(type) {
			case int64:
				f = float64(v)
			case float64:
				f = v
			default:
				assert.FailNow(t, "unexpected parsed type %T", a)
			}
			actual, err := FormatNumber(f, tc.prec, tc.base)
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
