package lexer

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
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

// NextRuneFrom reads the next rune from the given io.Reader and returns it along with the number of bytes read and any error encountered.
//
// Parameters:
// - reader: The io.Reader from which to read the rune.
//
// Returns:
// - ret: The rune read from the reader.
// - w: The number of bytes read from the reader.
// - err: An error if there was a problem reading the rune.
func NextRuneFrom(reader io.Reader) (data []byte, r rune, err error) {
	fetch := newFetcher(reader)
	defer fetch.Reset()
	for i := 1; i < 5; i++ {
		_, err = fetch.Fetch(i)
		if i == 1 && errors.Is(err, io.EOF) {
			break
		}
		decoded, _ := utf8.DecodeRune(fetch.Buffer())
		if err != nil || decoded != utf8.RuneError {
			r = decoded
			break
		}
	}
	data = fetch.Buffer()
	return
}
