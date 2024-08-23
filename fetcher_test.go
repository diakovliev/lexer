package lexer

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetcher(t *testing.T) {

	type expectation struct {
		bytes []byte
		size  int
		n     int
		err   error
	}

	type testCase struct {
		name         string
		reader       io.Reader
		expectations []expectation
	}

	testCases := []testCase{
		{
			name:   "read buffer",
			reader: bytes.NewBuffer([]byte{1, 2, 3, 4}),
			expectations: []expectation{
				{
					bytes: []byte{1},
					size:  1,
					n:     1,
				},
				{
					bytes: []byte{1, 2},
					size:  2,
					n:     2,
				},
				{
					bytes: []byte{1, 2, 3},
					size:  3,
					n:     3,
				},
				{
					bytes: []byte{1, 2, 3, 4},
					size:  4,
					n:     4,
				},
				{
					bytes: []byte{1, 2, 3, 4},
					size:  5,
					n:     4,
					err:   io.EOF,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pr := newFetcher(tc.reader)
			for i, exp := range tc.expectations {
				t.Run(fmt.Sprintf("%s-%d", tc.name, i), func(t *testing.T) {
					n, err := pr.Fetch(exp.size)
					assert.Equal(t, exp.n, n)
					if exp.err != nil {
						assert.Error(t, err)
						assert.ErrorIs(t, err, exp.err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, exp.bytes, pr.Buffer())
					}
				})
			}
		})
	}
}
