package xio

import "io"

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
		Begin() *Tx
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

	// State is the interface that groups the methods for IO state manipulation.
	State interface {
		Read
		Unread
		NextByte
		NextRune
		Has
		Data
	}

	// Source is the interface that groups the methods for IO source manipulation.
	// Use New to obtain new instance of Xio which is implements Source interface.
	Source interface {
		Has
		Begin
	}

	// Transaction is the interface that groups the methods for transaction manipulation.
	Transaction interface {
		Begin
		Commit
		Rollback
	}
)

// AsTransaction converts the given State to a Transaction if it possible.
// If the given State is not a Transaction it panics.
func AsTransaction(state State) (tx Transaction) {
	var i any = state
	tx, ok := i.(Transaction)
	if !ok {
		panic("not a Transaction")
	}
	return
}
