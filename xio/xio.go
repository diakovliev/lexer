package xio

import (
	"errors"
	"io"
	"math"

	"github.com/diakovliev/lexer/common"
)

type (
	// Xio is a buffered reader that allows to read from the buffer and rollback reads.
	// It implements Source interface.
	Xio struct {
		logger common.Logger
		reader io.Reader
		buffer *buffer
		pos    int64 // buffer position
		offset int64 // current position in the reader, used for transactions and truncates
		tx     *Tx
	}
)

// New creates new Xoi instance.
// The returned reader is buffered and can be used to rollback reads.
func New(logger common.Logger, r io.Reader) *Xio {
	return &Xio{
		logger: logger,
		reader: r,
		buffer: newBuffer([]byte{}),
		pos:    0,
		offset: 0,
	}
}

// Begin starts a new transaction for reading from the buffered reader.
func (r *Xio) Begin() (ret *Tx) {
	if r.tx != nil {
		r.logger.Fatal("too many transactions, Xio supports only one active transaction")
	}
	r.tx = newTx(r.logger, r, r.offset)
	ret = r.tx
	return
}

func (r *Xio) resetTx() {
	r.tx = nil
}

func (r Xio) len() (ret int) {
	ret = int(r.pos) + r.buffer.Len()
	return
}

func (r Xio) copyTo(pos int64, out []byte) (n int, err error) {
	start := int(pos)
	end := r.len()
	if int(pos)+len(out) < end {
		end = int(pos) + len(out)
	}
	data, err := r.Range(start, end)
	if err != nil {
		r.logger.Fatal("data range error: %s", err)
	}
	if n = copy(out, data); n != end-start {
		r.logger.Fatal("copied bytes count: %d != %d", n, end-start)
	}
	return
}

// Has returns true if the reader has more data to read.
func (r Xio) Has() (ret bool) {
	n, _ := r.Fetch(1)
	ret = n == 1
	return
}

// Range returns a slice of the buffered data between from and to positions.
func (r Xio) Range(from, to int) (out []byte, err error) {
	start := from - int(r.pos)
	// check bounds
	if start < 0 || start > r.buffer.Len() {
		r.logger.Fatal("out of bounds")
	}
	if math.MaxInt == to {
		out = r.buffer.Bytes()[start:]
	} else {
		end := to - int(r.pos)
		// check bounds
		if end < 0 || end > r.buffer.Len() || end < start {
			r.logger.Fatal("out of bounds")
		}
		out = r.buffer.Bytes()[start:end]
	}
	return
}

// Update updates the reader offset.
func (r *Xio) Update(offset int64) {
	r.offset = offset
}

// Truncate truncates the buffer from left up to the given position.
func (r *Xio) Truncate(pos int64) (err error) {
	if pos <= r.pos {
		return
	}
	// Do not truncate not "commit" by transaction data.
	// This is protection against incorrect position update
	// inside transaction implementation. Transaction
	// must update its position before calling truncate.
	if pos > r.offset {
		r.logger.Fatal("out of bounds")
	}
	data, err := r.Range(int(pos), math.MaxInt)
	if err != nil {
		r.logger.Fatal("data range error: %s", err)
	}
	newBuffer := newBuffer(data)
	r.buffer = newBuffer
	r.pos = pos
	return
}

// Fetch fetches data from the reader and appends it to the buffer if it needed.
// Fetch ensures that the buffer contains at least `size` bytes from the reader.
func (r Xio) Fetch(size int64) (n int64, err error) {
	if size <= 0 {
		return
	}
	r.buffer.Grow(int(size))
	n, err = io.CopyN(r.buffer, r.reader, size)
	return
}

// ReadAt reads from the buffered reader from given position and returns the number of bytes read.
func (r Xio) ReadAt(pos int64, out []byte) (n int, err error) {
	if pos < r.pos || pos < r.offset {
		r.logger.Fatal("out of bounds")
	}
	end := int(pos) + len(out)
	if r.len() >= end {
		if n, err = r.copyTo(pos, out); err != nil {
			r.logger.Fatal("copy to error: %s", err)
		}
		return
	}
	if _, err = r.Fetch(int64(end - r.len())); err != nil && !errors.Is(err, io.EOF) {
		r.logger.Error("fetch error: %s", err)
		return
	}
	// We need separate error variable to preserve original fetch error
	// in particular case of io.EOF.
	var copyErr error
	if n, copyErr = r.copyTo(pos, out); copyErr != nil {
		r.logger.Fatal("copy to error: %s", copyErr)
	}
	return
}
