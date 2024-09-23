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
		t.logger.Fatal("too many transactions, Tx supports only one active child transaction")
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
		t.logger.Fatal("transaction already complete")
	}
	// all children must be completed before committing the parent
	if t.tx != nil && t.tx.offset != -1 {
		t.logger.Fatal("child transaction is not complete")
	}
	if t.parent != nil {
		// update parent transaction position
		t.parent.update(t.offset)
		t.parent.resetTx()
	} else {
		// update reader position directly if no parent transaction exists
		t.reader.Update(t.offset)
		if err = t.reader.Truncate(t.offset); err != nil {
			t.logger.Fatal("truncate error: %s", err)
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
		t.logger.Fatal("transaction already complete")
		return
	}
	// rollback child transactions first
	if t.tx != nil && t.tx.offset != -1 {
		t.logger.Fatal("child transaction is not complete")
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
		t.logger.Fatal("transaction already complete")
		return
	}
	n, err = t.reader.ReadAt(t.offset, out)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Error("read error: %s", err)
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
		t.logger.Fatal("transaction already complete")
	}
	oldPos := t.offset
	newPos := oldPos - int64(t.lastN)
	t.offset = newPos
	t.lastN = 0
	return
}

// Data returns transaction data (reader data from offset to pos), updates pos and
// returns data position.
func (t *Tx) Data() (data []byte, pos int64, err error) {
	if t.offset == -1 {
		t.logger.Fatal("transaction already complete")
	}
	pos = t.pos
	data = make([]byte, t.offset-pos)
	n, err := t.reader.ReadAt(pos, data)
	if err != nil {
		data = nil
		return
	}
	if n != len(data) {
		t.logger.Fatal("data len error")
	}
	t.pos = t.offset
	return
}

// Has returns true if the transaction has data at pos.
func (t *Tx) Has() (ret bool) {
	if t.offset == -1 {
		t.logger.Fatal("transaction already complete")
	}
	data := make([]byte, 1)
	_, err := t.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Fatal("read error: %s", err)
	}
	ret = !errors.Is(err, io.EOF)
	if ret {
		if _, err = t.Unread(); err != nil {
			t.logger.Fatal("unread error: %s", err)
		}
	}
	return
}

func (t *Tx) nextBytes(size int) (data []byte, err error) {
	if t.offset == -1 {
		t.logger.Fatal("transaction already complete")
	}
	_, _ = t.reader.Fetch(utf8.UTFMax)
	data = make([]byte, size)
	n, err := t.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Fatal("read error: %s", err)
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
	if t.offset == -1 {
		t.logger.Fatal("transaction already complete")
		return
	}
	_, _ = t.reader.Fetch(utf8.UTFMax)
	data := make([]byte, utf8.UTFMax+1)
	offset := t.offset
	var i, n int
	for i = 1; i < utf8.UTFMax+1; i++ {
		n, err = t.reader.ReadAt(offset, data[:i])
		if err != nil && !errors.Is(err, io.EOF) {
			t.logger.Error("read error: %s", err)
			return
		}
		if n < i && errors.Is(err, io.EOF) {
			r = utf8.RuneError
			w = 0
			return
		}
		if r, w = utf8.DecodeRune(data[:i]); w != i {
			t.logger.Fatal("unexpected decoded rune width")
		}
		if w != utf8.RuneError {
			break
		}
	}
	// offset and lastN for Unread
	t.lastN = i
	t.offset += int64(t.lastN)
	return
}
