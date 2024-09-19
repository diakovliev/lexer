package common

import (
	"bytes"
	"errors"
	"io"
	"math"
)

var (
	// ErrOutOfBounds is returned when a read operation is out of bounds.
	ErrOutOfBounds = errors.New("out of bounds")
)

type (
	// Reader is a buffered reader that allows to read from the buffer and rollback reads.
	// Not thread-safe.
	Reader struct {
		logger    Logger
		reader    io.Reader
		buffer    *bytes.Buffer
		bufferPos int64 // buffer position
		pos       int64 // current position in the reader, used for transactions and truncates
		eof       bool
		tx        *Transaction
	}
)

// NewReader creates a new transaction reader from the given io.Reader.
// The returned reader is buffered and can be used to rollback reads.
func NewReader(logger Logger, r io.Reader) *Reader {
	return &Reader{
		logger:    logger,
		reader:    r,
		buffer:    bytes.NewBuffer(nil),
		bufferPos: 0,
		pos:       0,
	}
}

// Begin starts a new transaction for reading from the buffered reader.
func (r *Reader) Begin() (ret *Transaction) {
	if r.tx != nil {
		r.logger.Fatal("too many transactions, Reader supports only one active transaction")
		return
	}
	r.tx = newTransaction(r.logger, r, r.pos)
	r.logger.Trace("created transaction %p, reader: %p pos: %d", r.tx, r, r.pos)
	ret = r.tx
	return
}

func (r *Reader) resetTx() {
	r.tx = nil
}

// Has returns true if the reader has more data to read.
func (r Reader) Has() (ret bool) {
	data := make([]byte, 1)
	tx := r.Begin()
	_, err := tx.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		r.logger.Fatal("unexpected read error: %s", err)
	}
	ret = !errors.Is(err, io.EOF)
	if err = tx.Rollback(); err != nil {
		r.logger.Fatal("unexpected rollback error: %s", err)
	}
	return
}

func (r Reader) len() (ret int) {
	ret = int(r.bufferPos) + r.buffer.Len()
	return
}

func (r Reader) data(from, to int) (out []byte, err error) {
	start := from - int(r.bufferPos)
	// check bounds
	if start < 0 || start > r.buffer.Len() {
		err = ErrOutOfBounds
		return
	}
	if math.MaxInt == to {
		out = r.buffer.Bytes()[start:]
	} else {
		end := to - int(r.bufferPos)
		// check bounds
		if end < 0 || end > r.buffer.Len() || end < start {
			err = ErrOutOfBounds
			return
		}
		out = r.buffer.Bytes()[start:end]
	}
	return
}

func (r *Reader) update(pos int64, eof bool) {
	r.pos = pos
	r.eof = eof
}

// truncate truncates the buffer from left up to the given position.
func (r *Reader) truncate(pos int64) (err error) {
	if pos <= r.bufferPos {
		return
	}
	// Do not truncate not "commited" by transaction data.
	// This is protection against incorrect position update
	// inside transaction implementation. Transaction
	// must update its position before calling truncate.
	if pos > r.pos {
		err = ErrOutOfBounds
		r.logger.Error("out of bounds, pos=%d, r.pos=%d", pos, r.pos)
		return
	}
	data, err := r.data(int(pos), math.MaxInt)
	if err != nil {
		r.logger.Error("dataRange error: %s", err)
		return
	}
	newBuffer := bytes.NewBuffer(data)
	r.buffer = newBuffer
	r.bufferPos = pos
	return
}

func (r Reader) copyTo(pos int64, out []byte) (n int, err error) {
	start := int(pos)
	end := r.len()
	if int(pos)+len(out) < end {
		end = int(pos) + len(out)
	}
	data, err := r.data(start, end)
	if err != nil {
		r.logger.Error("data range error: %s", err)
		return
	}
	if n = copy(out, data); n != end-start {
		r.logger.Fatal("unexpected copied bytes count: %d != %d", n, end-start)
	}
	return
}

func (r Reader) fetch(size int64) (n int64, err error) {
	if size <= 0 {
		return
	}
	r.buffer.Grow(int(size))
	n, err = io.CopyN(r.buffer, r.reader, size)
	return
}

// readAt reads from the buffered reader from given position and returns the number of bytes read.
func (r Reader) readAt(pos int64, out []byte) (n int, err error) {
	if pos < r.bufferPos || pos < r.pos {
		err = ErrOutOfBounds
		r.logger.Error("out of bounds, r.pos=%d, pos=%d, r.bufferPos=%d", r.pos, pos, r.bufferPos)
		return
	}
	end := int(pos) + len(out)
	if r.len() >= end {
		if n, err = r.copyTo(pos, out); err != nil {
			r.logger.Error("copyTo error: %s", err)
		}
		return
	}
	if _, err = r.fetch(int64(end - r.len())); err != nil && !errors.Is(err, io.EOF) {
		r.logger.Error("fetch error: %s", err)
		return
	}
	// We need separate error variable to preserve original fetch
	// in particular case of io.EOF.
	var copyErr error
	if n, copyErr = r.copyTo(pos, out); copyErr != nil {
		err = copyErr
		r.logger.Error("copyTo error: %s", err)
	}
	return
}
