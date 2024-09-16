package lexer_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/diakovliev/lexer"
	"github.com/stretchr/testify/assert"
)

func TestNextByteFrom(t *testing.T) {

	type testCase struct {
		name   string
		reader io.Reader
		ret    byte
		w      int
		err    error
	}

	testCases := []testCase{
		{
			name:   "empty reader",
			reader: &bytes.Buffer{},
			ret:    0,
			w:      0,
			err:    io.EOF,
		},
		{
			name:   "single byte reader",
			reader: bytes.NewReader([]byte{1}),
			ret:    1,
			w:      1,
			err:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ret, w, err := lexer.NextByteFrom(tc.reader)
			assert.Equal(t, tc.ret, ret)
			assert.Equal(t, tc.w, w)
			if tc.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
