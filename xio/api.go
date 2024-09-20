package xio

import "io"

type (
	Unread interface {
		Unread() (n int, err error)
	}

	Begin interface {
		Begin() *Tx
	}

	Commit interface {
		Commit() error
	}

	Rollback interface {
		Rollback() error
	}

	NextByte interface {
		NextByte() (b byte, err error)
	}

	NextRune interface {
		NextRune() (r rune, w int, err error)
	}

	Data interface {
		NextByte
		NextRune
		Has() bool
		Data() (data []byte, pos int64, err error)
	}

	ReadUnreadData interface {
		Unread
		Data
		io.Reader
	}
)
