package xio

import (
	"io"

	"github.com/diakovliev/lexer/common"
)

type (
	// Read is a reference to io.Reader
	// It is read operation.
	Read = io.Reader

	// Unread unread the last read byte.
	// It undo last read operation.
	Unread interface {
		// Unread undo last read operation.
		Unread() (n int, err error)
	}

	// Begin begins new transaction.
	Begin interface {
		// Begin begins new transaction.
		Begin() common.Ref[State]
	}

	// Commit commits the transaction.
	Commit interface {
		// Commit commits the transaction.
		Commit() error
	}

	// Rollback rolls back the transaction.
	Rollback interface {
		// Rollback rolls back the transaction.
		Rollback() error
	}

	// NextByte returns next byte from the state.
	// It is read operation.
	NextByte interface {
		// NextByte returns next byte from the state.
		NextByte() (b byte, err error)
	}

	// NextRune returns next rune from the state.
	// It is read operation.
	NextRune interface {
		// NextRune returns next rune from the state.
		NextRune() (r rune, w int, err error)
	}

	// Has returns true if there are any bytes in state.
	// It is non-read operation.
	Has interface {
		// Has returns true if there are any bytes in state or in source.
		Has() bool
	}

	// Data extracts read data from the state, and advances state
	// to the returned amount of data.
	Data interface {
		// Data extracts read data from the state, and advances state
		Data() (data []byte, pos int64, err error)
	}

	// Buffer is the interface that groups the methods for IO buffer access.
	Buffer interface {
		// Buffer returns the buffer and its offset. It does not affect the state.
		Buffer() (ret []byte, offset int64, err error)
	}

	// State is the interface that groups the methods for IO state manipulation.
	State interface {
		Read
		Unread
		NextByte
		NextRune
		Has
		Data
		Buffer
	}

	// Tx is the interface that groups the methods for IO transaction manipulation.
	Tx interface {
		Commit
		Rollback
	}

	// Source is the interface that groups the methods for IO source manipulation.
	// Use New to obtain new instance of Xio which is implements Source interface.
	Source interface {
		Has
		Begin
		Buffer
	}
)

// AsSource converts the given State to a Source if it possible.
// If the given State is not a Source it panics.
func AsSource(state State) (source Source) {
	var i any = state
	source, ok := i.(Source)
	if !ok {
		panic("not a Source")
	}
	return
}

// AsTx converts the given State to a Tx if it possible.
func AsTx(state State) (tx Tx) {
	var i any = state
	tx, ok := i.(Tx)
	if !ok {
		panic("not a Tx")
	}
	return
}
