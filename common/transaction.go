package common

import (
	"errors"
	"io"
)

var (
	// ErrTransactionAlreadyCompleted is returned when a transaction has already been completed.
	ErrTransactionAlreadyCompleted = errors.New("transaction already completed")
	// ErrChildTransactionNotCompleted is returned when a child transaction has not been completed.
	ErrChildTransactionNotCompleted = errors.New("child transaction not completed")
	// ErrTransactionDataReadError is returned when there was an error reading from the buffered reader.
	ErrTransactionDataReadError = errors.New("transaction data read error")
)

type (
	// Transaction is a transaction for reading from the buffered reader.
	Transaction struct {
		logger      Logger
		reader      *Reader        // transaction reader
		parent      *Transaction   // parent transaction
		lastDataPos int64          // position of the last data returned by Data()
		pos         int64          // current position
		lastN       int            // last read bytes count
		children    []*Transaction // child transactions
		eof         bool           // end of file
	}
)

func newTransaction(logger Logger, reader *Reader, pos int64) (ret *Transaction) {
	ret = &Transaction{
		logger:      logger,
		reader:      reader,
		lastDataPos: pos,
		pos:         pos,
	}
	return
}

// Begin starts a new child transaction for reading from the buffered reader.
func (t *Transaction) Begin() (ret *Transaction) {
	ret = &Transaction{
		logger:      t.logger,
		reader:      t.reader,
		parent:      t,
		lastDataPos: t.pos,
		pos:         t.pos,
	}
	t.children = append(t.children, ret)
	t.logger.Trace("created transaction %p, parent: %p pos: %d", ret, ret.parent, ret.pos)
	return
}

func (t *Transaction) update(pos int64, eof bool) {
	t.pos = pos
	t.eof = eof
}

// Commit commits the transaction and returns the number of bytes read during the transaction.
// Commit will fail if any of the child transactions are not committed or rolled back.
func (t *Transaction) Commit() (n int, err error) {
	t.logger.Trace("=>> enter %p.Commit()", t)
	defer func() { t.logger.Trace("=>> leave %p.Commit() = n=%d err=%s", t, n, err) }()
	if t.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		t.logger.Error("%p: transaction already complete", t)
		return
	}
	// all children must be completed before committing the parent
	for _, child := range t.children {
		if child.pos != -1 {
			err = ErrChildTransactionNotCompleted
			t.logger.Error("%p: child transaction %p is not complete", t, child)
			return
		}
	}
	if t.parent != nil {
		// update parent transaction position
		n = int(t.pos - t.parent.pos)
		t.logger.Trace("%p: update parent transaction %p pos=%d->%d, eof=%v->%v", t, t.parent, t.parent.pos, t.pos, t.parent.eof, t.eof)
		t.parent.update(t.pos, t.eof)
	} else {
		// update reader position directly if no parent transaction exists
		n = int(t.pos - t.reader.pos)
		t.logger.Trace("%p: update reader %p pos=%d->%d, eof=%v->%v", t, t.reader, t.reader.pos, t.pos, t.reader.eof, t.eof)
		t.reader.update(t.pos, t.eof)
		if err = t.reader.truncate(t.pos); err != nil {
			t.logger.Fatal("%p: failed to truncate reader %p at pos=%d, err=%s", t, t.reader, t.pos, err)
			return
		}
		t.reader.resetTx()
	}
	// mark transaction as completed to prevent further use
	t.logger.Trace("%p: mark transaction as complete", t)
	t.lastDataPos = -1
	t.pos = -1
	t.lastN = 0
	return
}

// Rollback rolls back the transaction and returns an error if it was already committed or rolled back.
// Rollback will rollback all non completed children transactions if any.
func (t *Transaction) Rollback() (err error) {
	t.logger.Trace("=>> enter %p.Rollback()", t)
	defer func() { t.logger.Trace("=>> leave %p.Rollback() = err=%s", t, err) }()
	if t.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		t.logger.Error("%p: transaction already complete", t)
		return
	}
	// rollback all children transactions first
	for _, child := range t.children {
		if child.pos == -1 {
			continue
		}
		if err = child.Rollback(); err != nil {
			t.logger.Error("%p: rollback child transaction %p failed with error: %s", t, child, err)
			return
		}
	}
	if t.parent == nil {
		t.reader.resetTx()
	}
	// mark transaction as completed to prevent further use
	t.logger.Trace("%p: mark transaction as complete", t)
	t.lastDataPos = -1
	t.pos = -1
	t.eof = false
	t.lastN = 0
	return
}

// Read reads data from the transaction reader into a byte slice.
func (t *Transaction) Read(out []byte) (n int, err error) {
	t.logger.Trace("=>> enter %p.Read(%v)", t, out)
	defer func() { t.logger.Trace("=>> leave %p.Read(%v) = n=%d, err=%s", t, out, n, err) }()
	if t.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		t.logger.Error("%p: transaction already complete", t)
		return
	}
	n, err = t.reader.readAt(t.pos, out)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Error("%p: read error: %s", t, err)
		return
	}
	if errors.Is(err, io.EOF) {
		t.eof = true
		t.logger.Trace("eof=true")
	}
	t.lastN = n
	t.pos += int64(t.lastN)
	t.logger.Trace("%p: lastN=%d, pos=%d", t, t.lastN, t.pos)
	return
}

// Unread undoes the last Read call. It will return the transaction reader to the position
// it was at before the last Read call. If the transaction has already been committed or
// rolled back, this function has no effect.
func (t *Transaction) Unread() (n int, err error) {
	t.logger.Trace("=>> enter %p.Unread()", t)
	defer func() { t.logger.Trace("=>> leave %p.Unread()", t) }()
	if t.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		t.logger.Error("%p: transaction already complete", t)
		return
	}
	oldPos := t.pos
	newPos := oldPos - int64(t.lastN)
	t.logger.Trace("%p: unread, lastN=%d old pos=%d, new pos=%d", t, t.lastN, oldPos, newPos)
	t.pos = newPos
	t.eof = false
	t.lastN = 0
	return
}

// Data returns transaction data (reader data from lastDataPos to pos) and
// returns data position.
func (t *Transaction) Data() (pos int64, data []byte, err error) {
	if t.pos == -1 {
		err = ErrTransactionAlreadyCompleted
		return
	}
	pos = t.lastDataPos
	data = make([]byte, t.pos-pos)
	n, err := t.reader.readAt(pos, data)
	if err != nil {
		data = nil
		return
	}
	if n != len(data) {
		err = ErrTransactionDataReadError
		data = nil
	}
	t.lastDataPos = t.pos
	return
}

// Has returns true if the transaction has data at pos.
func (t Transaction) Has() (ret bool) {
	data := make([]byte, 1)
	tx := t.Begin()
	_, err := tx.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		t.logger.Fatal("unexpected read error: %s", err)
	}
	ret = !errors.Is(err, io.EOF)
	if err = tx.Rollback(); err != nil {
		t.logger.Fatal("unexpected rollback error: %s", err)
	}
	return
}
