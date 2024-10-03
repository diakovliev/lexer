package iterator

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
)

// Iterator implements iterator on top of the lexer.
type Iterator[T any] struct {
	logger     common.Logger
	ctx        context.Context
	lexer      *lexer.Lexer[T]
	bufferCap  int
	buffer     chan *message.Message[T]
	historyLen int
	history    []*message.Message[T]
	// Error is the last error that occurred except io.EOF
	Error error
}

// New creates a new iterator instance
func New[T any](reader io.Reader, opts ...Option[T]) (ret *Iterator[T]) {
	ret = &Iterator[T]{
		logger:     logger.Nop(),
		ctx:        context.Background(),
		bufferCap:  0,
		historyLen: 0,
	}
	for _, opt := range opts {
		opt(ret)
	}
	ret.lexer = lexer.New(ret.logger, reader, message.DefaultFactory[T](), ret)
	ret.buffer = make(chan *message.Message[T], ret.bufferCap)
	if ret.historyLen > 0 {
		ret.history = make([]*message.Message[T], 0, ret.historyLen)
	}
	return
}

// With sets the state provider
func (i *Iterator[T]) With(fn state.Provider[T]) *Iterator[T] {
	i.lexer.With(fn)
	return i
}

// addToHistory adds a message to the history
func (i *Iterator[T]) addToHistory(msg *message.Message[T]) {
	if i.historyLen == 0 {
		return
	}
	i.history = append(i.history, msg)
	if len(i.history) > i.historyLen {
		i.history = i.history[1:]
	}
}

// Receive implements the Receiver interface
func (i *Iterator[T]) Receive(msgs []*message.Message[T]) (err error) {
	for _, msg := range msgs {
		i.buffer <- msg
	}
	return
}

// History returns the history of the iterator.
func (i *Iterator[T]) History() []*message.Message[T] {
	return i.history
}

// Prev returns a message from the history with the given back index.
// If the back index is not given, the last message is returned.
// If the history is empty, nil is returned.
func (i *Iterator[T]) Prev(backIndex ...int) *message.Message[T] {
	index := 0
	if len(backIndex) > 0 {
		common.AssertTrue(backIndex[0] < len(i.history), "invalid back index")
		common.AssertTrue(len(backIndex) == 1, "invalid back index")
		index = backIndex[0]
	}
	if len(i.history) == 0 {
		return nil
	}
	return i.history[len(i.history)-index-1]
}

// Next returns the next message from the iterator.
// If the iterator is done, it returns (nil, io.EOF).
// If an error occurred, it returns (nil, err).
func (i *Iterator[T]) Next(ctx context.Context) (msg *message.Message[T], err error) {
	select {
	case <-ctx.Done():
		err = ctx.Err()
	case msg = <-i.buffer:
	}
	common.AssertFalse(msg == nil && err == nil, "both msg and err are nil")
	return
}

// Run runs the lexer and yields messages to the iterator channel until the lexer is done or
// an error occurs. If an error occurs, it is sent to the errors channel and both channels are closed.
func (i *Iterator[T]) Run(ctx context.Context) (err error) {
	if err = i.lexer.Run(ctx); err != nil && !errors.Is(err, io.EOF) {
		i.Error = err
	}
	return
}

// Range runs the iterator and yields messages to the given function until the iterator is done or
// an error occurs. If the given function returns false, the iterator is stopped and the function
// returns. If an error occurs, it is stored in the `Error` field of the iterator and the function
// returns.
func (i *Iterator[T]) Range(yield func(*message.Message[T]) bool) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(i.ctx)
	go func() {
		if err := i.Run(ctx); err != nil && !errors.Is(err, io.EOF) {
			i.Error = err
		}
		cancel()
		wg.Done()
	}()
	for {
		msg, err := i.Next(ctx)
		if err != nil && (errors.Is(err, io.EOF) || errors.Is(err, context.Canceled)) {
			break
		}
		if err != nil {
			common.AssertUnreachable("unexpected error: %s", err)
			break
		}
		if !yield(msg) {
			break
		}
		i.addToHistory(msg)
	}
	wg.Wait()
	close(i.buffer)
}
