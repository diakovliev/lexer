package common

import "io"

type (
	Unread interface {
		Unread() (n int, err error)
	}

	Tx interface {
		Begin() *Transaction
		Commit() (int, error)
		Rollback() error
	}

	Data interface {
		Has() bool
		Data() (pos int64, data []byte, err error)
	}

	ReadUnreadData interface {
		Unread
		Data
		io.Reader
	}

	Receiver[T any] func(Message[T]) error
)

func Dispose[T any](Message[T]) error {
	return nil
}
