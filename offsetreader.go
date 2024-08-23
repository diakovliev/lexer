package lexer

import (
	"io"
	"sync"
)

// OffsetReader is a reader that keeps track of the offset
type OffsetReader struct {
	sync.RWMutex
	offset int
	reader io.Reader
}

// NewOffsetReader creates a new OffsetReader
func NewOffsetReader(reader io.Reader) *OffsetReader {
	return &OffsetReader{
		offset: 0,
		reader: reader,
	}
}

// Offset returns the current offset
func (r *OffsetReader) Offset() int {
	r.RLock()
	defer r.RUnlock()
	return r.offset
}

// Read implements the io.Reader interface
func (r *OffsetReader) Read(p []byte) (n int, err error) {
	r.Lock()
	defer r.Unlock()
	n, err = r.reader.Read(p)
	r.offset += n
	return
}
