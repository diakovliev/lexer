package xio

import (
	"errors"
	"io"
	"unicode/utf8"

	"github.com/diakovliev/lexer/common"
)

type (
	// state is a transaction for reading from the buffered reader.
	// state implements State and Transaction interfaces.
	state struct {
		logger common.Logger
		reader *Xio   // transaction reader
		parent *state // parent transaction
		pos    int64  // position of the last data returned by Data()
		offset int64  // current position
		lastN  int    // last read bytes count
		tx     *state // child transactions
	}
)

func newState(logger common.Logger, reader *Xio, pos int64) (ret *state) {
	ret = &state{
		logger: logger,
		reader: reader,
		pos:    pos,
		offset: pos,
	}
	return
}

// Begin starts a child transaction.
func (s *state) Begin() (ret common.IfaceRef[State]) {
	if s.tx != nil {
		s.logger.Fatal("too many transactions, Tx supports only one active child transaction")
	}
	s.tx = &state{
		logger: s.logger,
		reader: s.reader,
		parent: s,
		pos:    s.offset,
		offset: s.offset,
	}
	ret = common.Ref[State](s.tx)
	return
}

func (s *state) resetTx() {
	s.tx = nil
}

func (s *state) update(offset int64) {
	s.offset = offset
}

func (s *state) reset() {
	s.pos = -1
	s.offset = -1
	s.lastN = 0
}

// Commit commits the transaction and returns the number of bytes read during the transaction.
// Commit will fail if any of the child transactions are not committed or rolled back.
func (s *state) Commit() (err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
	}
	// all children must be completed before committing the parent
	if s.tx != nil && s.tx.offset != -1 {
		s.logger.Fatal("child transaction is not complete")
	}
	if s.parent != nil {
		// update parent transaction position
		s.parent.update(s.offset)
		s.parent.resetTx()
	} else {
		// update reader position directly if no parent transaction exists
		s.reader.Update(s.offset)
		if err = s.reader.Truncate(s.offset); err != nil {
			s.logger.Fatal("truncate error: %s", err)
		}
		s.reader.resetTx()
	}
	s.reset()
	return
}

// Rollback rolls back the transaction and returns an error if it was already committed or rolled back.
// Rollback will rollback all non completed children transactions if any.
func (s *state) Rollback() (err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
		return
	}
	// rollback child transactions first
	if s.tx != nil && s.tx.offset != -1 {
		s.logger.Fatal("child transaction is not complete")
	}
	if s.parent != nil {
		s.parent.resetTx()
	} else {
		s.reader.resetTx()
	}
	s.reset()
	return
}

// Read reads data from the transaction reader into a byte slice.
func (s *state) Read(out []byte) (n int, err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
		return
	}
	n, err = s.reader.ReadAt(s.offset, out)
	if err != nil && !errors.Is(err, io.EOF) {
		s.logger.Error("read error: %s", err)
		return
	}
	s.lastN = n
	s.offset += int64(s.lastN)
	return
}

// Unread undoes the last Read call. It will return the transaction reader to the position
// it was at before the last Read call. If the transaction has already been committed or
// rolled back, this function has no effect.
func (s *state) Unread() (n int, err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
	}
	oldPos := s.offset
	newPos := oldPos - int64(s.lastN)
	s.offset = newPos
	s.lastN = 0
	return
}

// Data returns transaction data (reader data from offset to pos), updates pos and
// returns data position.
func (s *state) Data() (data []byte, pos int64, err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
	}
	pos = s.pos
	data = make([]byte, s.offset-pos)
	n, err := s.reader.ReadAt(pos, data)
	if err != nil {
		data = nil
		return
	}
	if n != len(data) {
		s.logger.Fatal("data len error")
	}
	s.pos = s.offset
	return
}

// Has returns true if the transaction has data at pos.
func (s *state) Has() (ret bool) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
	}
	data := make([]byte, 1)
	_, err := s.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		s.logger.Fatal("read error: %s", err)
	}
	ret = !errors.Is(err, io.EOF)
	if ret {
		if _, err = s.Unread(); err != nil {
			s.logger.Fatal("unread error: %s", err)
		}
	}
	return
}

func (s *state) nextBytes(size int) (data []byte, err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
	}
	_, _ = s.reader.Fetch(utf8.UTFMax)
	data = make([]byte, size)
	n, err := s.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		s.logger.Fatal("read error: %s", err)
	}
	data = data[:n]
	// lastN is set by nextBytes.
	return
}

// NextByte implements NextByte interface.
func (s *state) NextByte() (b byte, err error) {
	data, err := s.nextBytes(1)
	if len(data) != 0 {
		b = data[0]
	}
	// lastN is set by Read inside nextBytes.
	return
}

// NextRune implements NextRune interface.
func (s *state) NextRune() (r rune, w int, err error) {
	if s.offset == -1 {
		s.logger.Fatal("transaction already complete")
		return
	}
	_, _ = s.reader.Fetch(utf8.UTFMax)
	data := make([]byte, utf8.UTFMax+1)
	offset := s.offset
	var i, n int
	for i = 1; i < utf8.UTFMax+1; i++ {
		n, err = s.reader.ReadAt(offset, data[:i])
		if err != nil && !errors.Is(err, io.EOF) {
			s.logger.Error("read error: %s", err)
			return
		}
		if n < i && errors.Is(err, io.EOF) {
			r = utf8.RuneError
			w = 0
			return
		}
		if r, w = utf8.DecodeRune(data[:i]); w != i {
			s.logger.Fatal("unexpected decoded rune width")
		}
		if w != utf8.RuneError {
			break
		}
	}
	// offset and lastN for Unread
	s.lastN = i
	s.offset += int64(s.lastN)
	return
}

// Buffer returns the buffer and its offset. It does not affect the state.
func (s state) Buffer() (ret []byte, offset int64, err error) {
	return s.reader.Buffer()
}
