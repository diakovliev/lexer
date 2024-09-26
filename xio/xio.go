package xio

import (
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
		tx     *state
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
func (r *Xio) Begin() (ret common.IfaceRef[State]) {
	common.AssertNilPtr(r.tx, "too many transactions, Xio supports only one active transaction")
	r.tx = newState(r.logger, r, r.offset)
	ret = common.Ref[State](r.tx)
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
	common.AssertNoError(err, "data range error")
	n = copy(out, data)
	common.AssertTrue(n == end-start, "unexpected copied bytes count")
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
	common.AssertFalse(start < 0 || start > r.buffer.Len(), "out of bounds")
	if math.MaxInt == to {
		out = r.buffer.Bytes()[start:]
		return
	}
	end := to - int(r.pos)
	common.AssertFalse(end < 0 || end > r.buffer.Len() || end < start, "out of bounds")
	out = r.buffer.Bytes()[start:end]
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
	// Do not truncate the data what are "commit" by transaction.
	// This is protection against incorrect position update
	// inside transaction implementation. Transaction
	// must update reader position before Truncate call.
	common.AssertFalse(pos > r.offset, "out of bounds")
	data, err := r.Range(int(pos), math.MaxInt)
	common.AssertNoError(err, "data range error")
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
	common.AssertFalse(pos < r.pos || pos < r.offset, "out of bounds")
	end := int(pos) + len(out)
	if r.len() >= end {
		n, err = r.copyTo(pos, out)
		common.AssertNoError(err, "copy to error")
		return
	}
	_, err = r.Fetch(int64(end - r.len()))
	common.AssertNoErrorOrIs(err, io.EOF, "fetch error")
	// We need separate error variable to preserve original fetch error
	// in particular case of io.EOF.
	var copyErr error
	n, copyErr = r.copyTo(pos, out)
	common.AssertNoError(copyErr, "copy to error")
	return
}

// Buffer returns the buffer and its offset. It does not affect the state.
func (r Xio) Buffer() (ret []byte, offset int64, err error) {
	offset = r.offset
	ret = r.buffer.Bytes()
	return
}
