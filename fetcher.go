package lexer

import (
	"bytes"
	"io"
)

// fetcher is a reader wrapper that fetches particular amount of data from an io.Reader.
type fetcher struct {
	reader io.Reader
	buffer bytes.Buffer
}

// newFetcher creates a new instance of fetcher with the given io.Reader.
//
// Parameters:
// - reader: The io.Reader to read data from.
//
// Returns:
// - *prefetcher: A pointer to the newly created prefetcher.
func newFetcher(reader io.Reader) *fetcher {
	return &fetcher{
		reader: reader,
	}
}

// Fetch reads data from the underlying reader and stores it in a buffer until the buffer size reaches the specified size.
//
// Parameters:
// - size: The target size of the buffer.
//
// Returns:
// - n: The current size of the buffer.
// - err: An error if there was a problem reading from the reader.
func (p *fetcher) Fetch(size int) (n int, err error) {
	n = p.buffer.Len()
	if n >= size {
		return
	}
	toRead := size - p.buffer.Len()
	in := make([]byte, toRead)
	read, err := p.reader.Read(in)
	if read > 0 {
		p.buffer.Grow(toRead)
		if _, writeErr := p.buffer.Write(in[:read]); writeErr != nil {
			return
		}
	}
	n = p.buffer.Len()
	return
}

// Buffer returns a slice of bytes from the buffer of the fetcher.
//
// It returns the underlying bytes slice of the buffer.
// The returned slice is a view of the buffer's bytes, so any changes made to the
// returned slice will be reflected in the buffer.
//
// Returns:
// - []byte: A slice of bytes from the buffer.
func (p *fetcher) Buffer() []byte {
	return p.buffer.Bytes()
}

// Len returns the number of bytes currently in the buffer of the fetcher.
//
// It returns an integer value representing the number of bytes in the buffer.
func (p *fetcher) Len() int {
	return p.buffer.Len()
}

// Reset resets the fetcher's buffer by calling the Reset method on the underlying bytes.Buffer.
//
// This method is used to clear the buffer and prepare it for reading new data from the underlying reader.
func (p *fetcher) Reset() {
	p.buffer.Reset()
}
