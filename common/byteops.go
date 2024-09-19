package common

import (
	"io"
)

// NextByteFrom reads a single byte from the given io.Reader and returns it along with the number of bytes read and any error encountered.
//
// Parameters:
// - reader: The io.Reader from which to read the byte.
//
// Returns:
// - ret: The byte read from the reader.
// - w: The number of bytes read from the reader.
// - err: An error if there was a problem reading the byte.
func NextByteFrom(reader io.Reader) (ret byte, w int, err error) {
	in := make([]byte, 1)
	_, err = reader.Read(in)
	if err != nil {
		return
	}
	ret = in[0]
	w = 1
	return
}
