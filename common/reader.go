package common

import (
	"bytes"
	"errors"
	"io"
	"math"
)

var (
	// ErrUnexpectedCopiedBytesCountError is returned when a copy operation fails unexpectedly.
	ErrUnexpectedCopiedBytesCountError = errors.New("unexpected copied bytes count error")
	// ErrUnexpectedWrittenBytesCountError is returned when a write operation fails unexpectedly.
	ErrUnexpectedWrittenBytesCountError = errors.New("unexpected written bytes count error")
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
		activeTx  *Transaction
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
	if r.activeTx != nil {
		r.logger.Fatal("too many transactions, Reader supports only one active transaction")
		return
	}
	r.activeTx = newTransaction(r.logger, r, r.pos)
	r.logger.Trace("created transaction %p, reader: %p pos: %d", r.activeTx, r, r.pos)
	ret = r.activeTx
	return
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

func (r Reader) dataLen() (ret int) {
	ret = int(r.bufferPos) + r.buffer.Len()
	return
}

func (r Reader) dataRange(from, to int) (out []byte, err error) {
	buffStartPos := from - int(r.bufferPos)
	// check bounds
	if buffStartPos < 0 || buffStartPos > r.buffer.Len() {
		err = ErrOutOfBounds
		return
	}
	if math.MaxInt == to {
		out = r.buffer.Bytes()[buffStartPos:]
	} else {
		buffEndPos := to - int(r.bufferPos)
		// check bounds
		if buffEndPos < 0 || buffEndPos > r.buffer.Len() || buffEndPos < buffStartPos {
			err = ErrOutOfBounds
			return
		}
		out = r.buffer.Bytes()[buffStartPos:buffEndPos]
	}
	return
}

// truncate truncates the buffer from left up to the given position.
func (r *Reader) truncate(pos int64) (err error) {
	r.logger.Trace("=>> enter truncate(%v)", pos)
	defer func() { r.logger.Trace("<<= leave truncate(%v) = err=%s", pos, err) }()

	// check lower bound
	if pos > r.pos {
		err = ErrOutOfBounds
		r.logger.Error("out of bounds, pos=%d, r.pos=%d", pos, r.pos)
		return
	}

	oldBufferLen := r.buffer.Len()

	// create a new buffer with the remaining data
	newBuffer := bytes.NewBuffer(nil)
	newBuffer.Grow(r.dataLen() - int(pos))
	data, err := r.dataRange(int(pos), math.MaxInt)
	if _, err = newBuffer.Write(data); err != nil {
		r.logger.Error("unexpected write error: %s", err)
		return
	}

	// set new buffer and update r.bufferPos accordingly
	r.buffer = newBuffer
	r.bufferPos = pos

	r.logger.Trace("truncated: %d", oldBufferLen-r.buffer.Len())

	return
}

// readAt reads from the buffered reader from given position and returns the number of bytes read.
func (r Reader) readAt(pos int64, out []byte) (n int, err error) {
	r.logger.Trace("=>> enter readAt(%v, %v)", pos, out)
	defer func() { r.logger.Trace("<<= leave readAt(%v, %v) = n=%d, err=%s", pos, out, n, err) }()

	// check lower bound
	if pos < r.pos {
		err = ErrOutOfBounds
		r.logger.Error("out of bounds, pos=%d, r.pos=%d", pos, r.pos)
		return
	}

	futurePos := int(pos) + len(out)
	r.logger.Trace("futurePos=%d, tr.dataLen()=%d", futurePos, r.dataLen())
	if r.dataLen() >= futurePos {
		data, rangeErr := r.dataRange(int(pos), futurePos)
		if rangeErr != nil {
			err = rangeErr
			r.logger.Error("dataRange error: %s", rangeErr)
			return
		}
		if n = copy(out, data); n != len(out) {
			err = ErrUnexpectedCopiedBytesCountError
			r.logger.Error("unexpected copied bytes count: %d != %d", n, len(out))
		}
		return
	}

	// prepare counters
	requested := futurePos - r.dataLen()
	if requested == 0 {
		return
	}

	// increase buffer capacity to at least toRead bytes
	r.logger.Trace("tr.buffer.Grow(%d)", requested)
	r.buffer.Grow(requested)

	// fetch data from the underlying reader
	in := make([]byte, requested)
	fetched := 0
	for requested > 0 && fetched < requested {
		r.logger.Trace("requested=%d, completed=%d", requested, fetched)
		read, readErr := r.reader.Read(in[0 : requested-fetched])
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			err = readErr
			r.logger.Error("read error: %s", readErr)
			return
		}
		if read > 0 {
			written, writeErr := r.buffer.Write(in[:read])
			if writeErr != nil {
				err = writeErr
				r.logger.Error("write error: %s", writeErr)
				return
			}
			if written != read {
				err = ErrUnexpectedWrittenBytesCountError
				r.logger.Error("unexpected written bytes count: %d != %d", written, read)
				return
			}
		}
		fetched += read
		r.logger.Trace("requested=%d, completed=%d", requested, fetched)
		if errors.Is(readErr, io.EOF) {
			err = readErr
			r.logger.Trace("EOF")
			break
		}
	}

	// copy data from buffer to the destination slice
	copyFrom := int(pos)
	copyTo := r.dataLen()
	if int(pos)+len(out) < copyTo {
		copyTo = int(pos) + len(out)
	}
	toCopy := copyTo - copyFrom

	data, dataErr := r.dataRange(copyFrom, copyTo)
	if dataErr != nil {
		err = dataErr
		r.logger.Error("data range error: %s", err)
		return
	}
	r.logger.Trace("data len: %d, to copy: %d", len(data), toCopy)
	if n = copy(out, data); n != toCopy {
		// we are copied less than expected bytes
		err = ErrUnexpectedCopiedBytesCountError
		r.logger.Error("unexpected copied bytes count: %d != %d", n, len(out))
	}

	return
}
