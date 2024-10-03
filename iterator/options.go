package iterator

import (
	"context"

	"github.com/diakovliev/lexer/common"
)

// Option is an option that can be passed to the iterator constructor.
type Option[T any] func(*Iterator[T])

// WithLogger sets the logger to use for logging. If not set, it will default to
// the logger from the lexer.
func WithLogger[T any](logger common.Logger) Option[T] {
	return func(it *Iterator[T]) {
		it.logger = logger
	}
}

// WithHistoryLen sets the maximum number of items to keep in the history buffer.
// If the length is negative, the history buffer is disabled.
func WithHistoryLen[T any](len int) Option[T] {
	return func(it *Iterator[T]) {
		it.historyLen = len
	}
}

// WithBufferCapacity sets the capacity of the buffer channel used to pass
// messages to the user. If the buffer is full, the iterator will block until
// the user takes a message from the buffer. The default buffer capacity is 0,
// which means that the iterator will not block and will send messages to the
// user immediately. If the user does not take the message from the buffer in
// time, the message will be lost.
func WithBufferCapacity[T any](cap int) Option[T] {
	return func(it *Iterator[T]) {
		it.bufferCap = cap
	}
}

// WithContext sets the context to use for the iterator. If not set, it will
// default to the background context.
func WithContext[T any](ctx context.Context) Option[T] {
	return func(it *Iterator[T]) {
		it.ctx = ctx
	}
}
