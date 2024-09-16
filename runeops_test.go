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

// func TestNextRuneFrom(t *testing.T) {

// 	type testCase struct {
// 		name   string
// 		reader io.Reader
// 		ret    rune
// 		w      int
// 		err    error
// 	}

// 	// 'Valid 2 Octet Sequence' => "\xc3\xb1",
// 	// 'Invalid 2 Octet Sequence' => "\xc3\x28",
// 	// 'Invalid Sequence Identifier' => "\xa0\xa1",
// 	// 'Valid 3 Octet Sequence' => "\xe2\x82\xa1",
// 	// 'Invalid 3 Octet Sequence (in 2nd Octet)' => "\xe2\x28\xa1",
// 	// 'Invalid 3 Octet Sequence (in 3rd Octet)' => "\xe2\x82\x28",
// 	// 'Valid 4 Octet Sequence' => "\xf0\x90\x8c\xbc",
// 	// 'Invalid 4 Octet Sequence (in 2nd Octet)' => "\xf0\x28\x8c\xbc",
// 	// 'Invalid 4 Octet Sequence (in 3rd Octet)' => "\xf0\x90\x28\xbc",
// 	// 'Invalid 4 Octet Sequence (in 4th Octet)' => "\xf0\x28\x8c\x28",
// 	testCases := []testCase{
// 		{
// 			name:   "valid 2 octet sequence",
// 			reader: bytes.NewBufferString("\xc3\xb1"),
// 			ret:    241,
// 			w:      2,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 2 octet sequence eof",
// 			reader: bytes.NewBufferString("\xc3\x28"),
// 			ret:    utf8.RuneError,
// 			w:      2,
// 			err:    io.EOF,
// 		},
// 		{
// 			name:   "invalid 2 octet sequence",
// 			reader: bytes.NewBufferString("\xc3\x28test"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid sequence identifier",
// 			reader: bytes.NewBufferString("\xa0\xa1"),
// 			ret:    utf8.RuneError,
// 			w:      2,
// 			err:    io.EOF,
// 		},
// 		{
// 			name:   "valid 3 octet sequence",
// 			reader: bytes.NewBufferString("\xe2\x82\xa1"),
// 			ret:    8353,
// 			w:      3,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 3 octet sequence (in 2nd octet) eof",
// 			reader: bytes.NewBufferString("\xe2\x28\xa1"),
// 			ret:    utf8.RuneError,
// 			w:      3,
// 			err:    io.EOF,
// 		},
// 		{
// 			name:   "invalid 3 octet sequence (in 3rd octet) eof",
// 			reader: bytes.NewBufferString("\xe2\x82\x28"),
// 			ret:    utf8.RuneError,
// 			w:      3,
// 			err:    io.EOF,
// 		},
// 		{
// 			name:   "invalid 3 octet sequence (in 3rd octet)",
// 			reader: bytes.NewBufferString("\xe2\x82\x28test"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "valid 4 octet sequence",
// 			reader: bytes.NewBufferString("\xf0\x90\x8c\xbc"),
// 			ret:    66364,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 4 octet sequence (in 2nd octet)",
// 			reader: bytes.NewBufferString("\xf0\x28\x8c\xbc"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 4 octet sequence (in 3rd octet)",
// 			reader: bytes.NewBufferString("\xf0\x90\x28\xbc"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 4 octet sequence (in 4th octet)",
// 			reader: bytes.NewBufferString("\xf0\x28\x8c\x28"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "invalid 4 octet sequence (in 4th octet)",
// 			reader: bytes.NewBufferString("\xf0\x28\x8c\x28test"),
// 			ret:    utf8.RuneError,
// 			w:      4,
// 			err:    nil,
// 		},
// 		{
// 			name:   "empty reader",
// 			reader: &bytes.Buffer{},
// 			ret:    0,
// 			w:      0,
// 			err:    io.EOF,
// 		},
// 		{
// 			name:   "single byte reader",
// 			reader: bytes.NewBufferString("1"),
// 			ret:    '1',
// 			w:      1,
// 			err:    nil,
// 		},
// 		{
// 			name:   "multi byte reader",
// 			reader: bytes.NewBufferString("11"),
// 			ret:    '1',
// 			w:      1,
// 			err:    nil,
// 		},
// 		{
// 			name:   "multi byte reader",
// 			reader: bytes.NewBufferString("111"),
// 			ret:    '1',
// 			w:      1,
// 			err:    nil,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ret, w, err := NextRuneFrom(tc.reader)
// 			assert.Equal(t, tc.ret, ret)
// 			assert.Equal(t, tc.w, w)
// 			if tc.err != nil {
// 				assert.Error(t, err)
// 				assert.ErrorIs(t, err, tc.err)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 		})
// 	}
// }
