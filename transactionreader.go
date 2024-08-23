package lexer

import (
	"bytes"
	"errors"
	"io"
)

var (
	// ErrTransactionAlreadyCompleted is returned when a transaction has already been completed.
	ErrTransactionAlreadyCompleted = errors.New("transaction already completed")
	// ErrChildTransactionNotCompleted is returned when a child transaction has not been completed.
	ErrChildTransactionNotCompleted = errors.New("child transaction not completed")
	// ErrUnexpectedCopiedBytesCountError is returned when a copy operation fails unexpectedly.
	ErrUnexpectedCopiedBytesCountError = errors.New("unexpected copied bytes count error")
	// ErrUnexpectedWrittenBytesCountError is returned when a write operation fails unexpectedly.
	ErrUnexpectedWrittenBytesCountError = errors.New("unexpected written bytes count error")
)

type (
	// TransactionReader is a buffered reader that allows to read from the buffer and rollback reads.
	TransactionReader struct {
		reader io.Reader
		buffer *bytes.Buffer
		pos    int64
		eof    bool
	}

	// ReaderTransaction is a transaction for reading from the buffered reader.
	ReaderTransaction struct {
		reader   *TransactionReader   // transaction reader
		parent   *ReaderTransaction   // parent transaction
		pos      int64                // current position
		lastN    int                  // last read bytes count
		children []*ReaderTransaction // child transactions
		eof      bool                 // end of file
	}
)

// NewTransactionReader creates a new transaction reader from the given io.Reader.
// The returned reader is buffered and can be used to rollback reads.
func NewTransactionReader(r io.Reader) *TransactionReader {
	return &TransactionReader{
		reader: r,
		buffer: bytes.NewBuffer([]byte{}),
		pos:    0,
	}
}

// Begin starts a new transaction for reading from the buffered reader.
func (tr *TransactionReader) Begin() (rt *ReaderTransaction) {
	rt = &ReaderTransaction{
		reader: tr,
		pos:    tr.pos,
	}
	return
}

// Pos returns the current position of the reader.
func (tr TransactionReader) Pos() int64 {
	return tr.pos
}

// EOF returns true if the reader has reached EOF.
func (tr TransactionReader) EOF() bool {
	return tr.eof
}

// Buffer returns the underlying buffer of the transaction reader.
func (tr TransactionReader) Buffer() *bytes.Buffer {
	return tr.buffer
}

// readAt reads from the buffered reader from given position and returns the number of bytes read.
func (tr TransactionReader) readAt(pos int64, out []byte) (n int, err error) {
	futurePos := int(pos) + len(out)
	if tr.buffer.Len() >= futurePos {
		if n = copy(out, tr.buffer.Bytes()[pos:futurePos]); n != len(out) {
			// we are copied less than expected bytes
			err = ErrUnexpectedCopiedBytesCountError
		}
		return
	}
	requested := futurePos - tr.buffer.Len()
	if requested == 0 {
		return
	}
	// increase buffer capacity to at least toRead bytes
	tr.buffer.Grow(requested)
	// read requested data
	completed := 0
	in := make([]byte, requested)
	for requested > 0 && completed < requested {
		read, readErr := tr.reader.Read(in[0 : requested-completed])
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			err = readErr
			return
		}
		if read > 0 {
			written, writeErr := tr.buffer.Write(in[:read])
			if writeErr != nil {
				err = writeErr
				return
			}
			if written != read {
				err = ErrUnexpectedWrittenBytesCountError
				return
			}
		}
		completed += read
		if errors.Is(readErr, io.EOF) {
			err = readErr
			break
		}
	}
	if n = copy(out, tr.buffer.Bytes()[pos:pos+int64(completed)]); n != completed {
		// we are copied less than expected bytes
		err = ErrUnexpectedCopiedBytesCountError
	}
	return
}

// Begin starts a new child transaction for reading from the buffered reader.
func (rt *ReaderTransaction) Begin() (ret *ReaderTransaction) {
	ret = &ReaderTransaction{
		reader: rt.reader,
		parent: rt,
		pos:    rt.pos,
	}
	rt.children = append(rt.children, ret)
	return
}

// Parent returns the parent transaction of the current transaction.
// If the transaction is the root transaction, the function returns nil.
func (rt *ReaderTransaction) Parent() *ReaderTransaction {
	return rt.parent
}

// Commit commits the transaction and returns the number of bytes read during the transaction.
// Commit will fail if any of the child transactions are not committed or rolled back.
func (rt *ReaderTransaction) Commit() (n int, err error) {
	if rt.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		return
	}
	// all children must be completed before committing the parent
	for _, child := range rt.children {
		if child.pos != -1 {
			err = ErrChildTransactionNotCompleted
			return
		}
	}
	if rt.parent != nil {
		// update parent transaction position
		n = int(rt.pos - rt.parent.pos)
		rt.parent.pos = rt.pos
		rt.parent.eof = rt.eof
	} else {
		// update reader position directly if no parent transaction exists
		n = int(rt.pos - rt.reader.pos)
		rt.reader.pos = rt.pos
		rt.reader.eof = rt.eof
	}
	// mark transaction as completed to prevent further use
	rt.pos = -1
	rt.lastN = 0
	return
}

// Rollback rolls back the transaction and returns an error if it was already committed or rolled back.
// Rollback will rollback all non completed children transactions if any.
func (rt *ReaderTransaction) Rollback() (err error) {
	if rt.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		return
	}
	// rollback all children transactions first
	for _, child := range rt.children {
		if child.pos == -1 {
			continue
		}
		if err = child.Rollback(); err != nil {
			return
		}
	}
	// mark transaction as completed to prevent further use
	rt.pos = -1
	rt.lastN = 0
	return
}

// Pos returns the current position of the transaction reader.
func (rt ReaderTransaction) Pos() int64 {
	return rt.pos
}

func (rt *ReaderTransaction) WithPosition(pos int64) *ReaderTransaction {
	rt.pos = pos
	return rt
}

// EOF returns true if the transaction reader has reached the end of file.
func (rt ReaderTransaction) EOF() bool {
	return rt.eof
}

// Buffer returns the underlying buffer of the transaction reader.
func (rt ReaderTransaction) Buffer() *bytes.Buffer {
	return rt.reader.Buffer()
}

// Read reads data from the transaction reader into a byte slice.
func (rt *ReaderTransaction) Read(out []byte) (n int, err error) {
	if rt.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		return
	}
	n, err = rt.reader.readAt(rt.pos, out)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, io.EOF) {
		rt.eof = true
	}
	rt.lastN = n
	rt.pos += int64(rt.lastN)
	return
}

// Unread undoes the last Read call. It will return the transaction reader to the position
// it was at before the last Read call. If the transaction has already been committed or
// rolled back, this function has no effect.
func (rt *ReaderTransaction) Unread() *ReaderTransaction {
	if rt.pos == -1 {
		return rt
	}
	rt.pos -= int64(rt.lastN)
	rt.lastN = 0
	return rt
}
