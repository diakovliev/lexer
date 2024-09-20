package xio

import (
	"errors"
	"io"
	"unicode/utf8"

	"github.com/diakovliev/lexer/common"
)

type (
	// Tx is a transaction for reading from the buffered reader.
	// Tx implements State and Transaction interfaces.
	Tx struct {
		logger common.Logger
		reader *Xio  // transaction reader
		parent *Tx   // parent transaction
		pos    int64 // position of the last data returned by Data()
		offset int64 // current position
		lastN  int   // last read bytes count
		tx     *Tx   // child transactions
	}
)

func newTx(logger common.Logger, reader *Xio, pos int64) (ret *Tx) {
	ret = &Tx{
		logger: logger,
		reader: reader,
		pos:    pos,
		offset: pos,
	}
	return
}

// Begin starts a child transaction.
func (t *Tx) Begin() (ret *Tx) {
	if t.tx != nil {
		t.logger.Fatal("%p: too many transactions, Transaction supports only one active child transaction", t)
	}
	ret = &Tx{
		logger: t.logger,
		reader: t.reader,
		parent: t,
		pos:    t.offset,
		offset: t.offset,
	}
	t.tx = ret
	return
}

func (t *Tx) resetTx() {
	t.tx = nil
}

func (t *Tx) update(offset int64) {
	t.offset = offset
}

func (t *Tx) reset() {
	t.pos = -1
	t.offset = -1
	t.lastN = 0
}

// Commit commits the transaction and returns the number of bytes read during the transaction.
// Commit will fail if any of the child transactions are not committed or rolled back.
func (t *Tx) Commit() (err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
	}
	// all children must be completed before committing the parent
	if t.tx != nil && t.tx.offset != -1 {
		t.logger.Fatal("%p: child transaction %p is not complete", t, t.tx)
	}
	if t.parent != nil {
		// update parent transaction position
		t.parent.update(t.offset)
		t.parent.resetTx()
	} else {
		// update reader position directly if no parent transaction exists
		t.reader.Update(t.offset)
		if err = t.reader.Truncate(t.offset); err != nil {
			t.logger.Fatal("%p: failed to truncate reader %p at pos=%d, err=%s", t, t.reader, t.offset, err)
		}
		t.reader.resetTx()
	}
	t.reset()
	return
}

// Rollback rolls back the transaction and returns an error if it was already committed or rolled back.
// Rollback will rollback all non completed children transactions if any.
func (t *Tx) Rollback() (err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
		return
	}
	// rollback child transactions first
	if t.tx != nil && t.tx.offset != -1 {
		t.logger.Fatal("%p: child transaction %p is not complete", t, t.tx)
	}
	if t.parent != nil {
		t.parent.resetTx()
	} else {
		t.reader.resetTx()
	}
	t.reset()
	return
}

// Read reads data from the transaction reader into a byte slice.
func (t *Tx) Read(out []byte) (n int, err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
		return
	}
	n, err = t.reader.ReadAt(t.offset, out)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Error("%p: read error: %s", t, err)
		return
	}
	t.lastN = n
	t.offset += int64(t.lastN)
	return
}

// Unread undoes the last Read call. It will return the transaction reader to the position
// it was at before the last Read call. If the transaction has already been committed or
// rolled back, this function has no effect.
func (t *Tx) Unread() (n int, err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
	}
	oldPos := t.offset
	newPos := oldPos - int64(t.lastN)
	t.offset = newPos
	t.lastN = 0
	return
}

// Data returns transaction data (reader data from lastDataPos to pos) and
// returns data position.
func (t *Tx) Data() (data []byte, pos int64, err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
	}
	pos = t.pos
	data = make([]byte, t.offset-pos)
	n, err := t.reader.ReadAt(pos, data)
	if err != nil {
		data = nil
		return
	}
	if n != len(data) {
		t.logger.Fatal("%p: data len: expected: %d, got: %d", t, n, len(data))
	}
	t.pos = t.offset
	return
}

// Has returns true if the transaction has data at pos.
func (t *Tx) Has() (ret bool) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
	}
	data := make([]byte, 1)
	_, err := t.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Fatal("%p: read error: %s", t, err)
	}
	ret = !errors.Is(err, io.EOF)
	if ret {
		if _, err = t.Unread(); err != nil {
			t.logger.Fatal("%p: unread error: %s", t, err)
		}
	}
	return
}

func (t *Tx) nextBytes(size int) (data []byte, err error) {
	if t.offset == -1 {
		t.logger.Fatal("%p: transaction already complete", t)
	}
	data = make([]byte, size)
	n, err := t.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Fatal("%p: read error: %s", t, err)
	}
	data = data[:n]
	// lastN is set by nextBytes.
	return
}

// NextByte implements NextByte interface.
func (t *Tx) NextByte() (b byte, err error) {
	data, err := t.nextBytes(1)
	if len(data) != 0 {
		b = data[0]
	}
	// lastN is set by Read inside nextBytes.
	return
}

// NextRune implements NextRune interface.
func (t *Tx) NextRune() (r rune, w int, err error) {
	for i := 1; i < utf8.UTFMax+1; i++ {
		tx := t.Begin()
		data, nextBytesErr := tx.nextBytes(i)
		if i == 1 && errors.Is(nextBytesErr, io.EOF) {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				t.logger.Fatal("%p: rollback error: %s", t, rollbackErr)
			}
			err = io.EOF
			break
		}
		decoded, decodedSize := utf8.DecodeRune(data)
		if decodedSize != i {
			t.logger.Fatal("%p: %d != %d", t, decodedSize, i)
		}
		if nextBytesErr != nil || decoded != utf8.RuneError {
			if commitErr := tx.Commit(); commitErr != nil {
				t.logger.Fatal("%p: commit error: %s", t, commitErr)
			}
			r = decoded
			w = len(data)
			// lastN for Unread
			t.lastN = w
			err = nextBytesErr
			break
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			t.logger.Fatal("%p: rollback error: %s", t, rollbackErr)
		}
	}
	return
}
