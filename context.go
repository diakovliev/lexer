package lexer

import (
	"errors"
	"io"
)

var (
	ErrBreak = errors.New("break")
)

type (
	ContextReceiver[T any] func(messages []Message[T])

	Context[T any] struct {
		parentTx *ReaderTransaction
		tx       *ReaderTransaction
		messages []Message[T]
		receiver ContextReceiver[T]
		Current  *Acceptor[T]
		Error    error
	}
)

func NewContext[T any](tx *ReaderTransaction, yeild ContextReceiver[T]) *Context[T] {
	return &Context[T]{
		parentTx: tx,
		messages: make([]Message[T], 0, 10),
		receiver: yeild,
	}
}

func (c *Context[T]) addMessage(message Message[T]) {
	c.messages = append(c.messages, message)
}

func (c *Context[T]) Fn(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).Fn(acceptFn)
	return c.Current
}

func (c *Context[T]) Until(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).Until(acceptFn)
	return c.Current
}

func (c *Context[T]) While(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).While(acceptFn)
	return c.Current
}

func (c *Context[T]) Count(count int) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).Count(count)
	return c.Current
}

func (c *Context[T]) String(match string) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).String(match)
	return c.Current
}

func (c *Context[T]) AnyString(matches ...string) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AnyString(matches...)
	return c.Current
}

func (c *Context[T]) AnyFn(acceptFns ...func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AnyFn(acceptFns...)
	return c.Current
}

func (c *Context[T]) OptionallyWhile(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).OptionallyWhile(acceptFn)
	return c.Current
}

func (c *Context[T]) OptionallyUntil(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).OptionallyUntil(acceptFn)
	return c.Current
}

func (c *Context[T]) If(ctxFn func(*Context[T])) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).If(ctxFn)
	return c.Current
}

func (c *Context[T]) Optionally(ctxFn func(*Context[T])) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).Optionally(ctxFn)
	return c.Current
}

func (c *Context[T]) complete() bool {
	if c.Error != nil && !errors.Is(c.Error, ErrBreak) {
		return true
	}
	if err := c.commit(); err != nil {
		c.SetError(err)
		return true
	}
	if errors.Is(c.Error, ErrBreak) {
		c.SetError(nil)
		return true
	}
	if c.tx == nil {
		c.tx = c.parentTx.Begin()
	}
	return false
}

func (c *Context[T]) Break() {
	c.SetError(ErrBreak)
}

func (c *Context[T]) SetError(err error) {
	c.Error = err
}

func (c *Context[T]) isEOF() (err error) {
	tx := c.parentTx.Begin()
	// try to read next byte
	if _, _, err = NextByteFrom(tx); err == nil {
		// if no error then rollback parent transaction
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			err = rollbackErr
		}
		return
	}
	if errors.Is(err, io.EOF) {
		// if EOF then commit parent transaction and return EOF
		if _, commitErr := tx.Commit(); commitErr != nil {
			err = commitErr
		}
	}
	// otherwise return error
	return
}

func (c *Context[T]) commit() (err error) {
	if c.tx == nil {
		return
	}
	if _, err = c.tx.Commit(); err != nil {
		c.Error = err
		return
	}
	c.receiver(c.messages)
	c.tx = nil
	c.messages = make([]Message[T], 0, 10)
	// if not EOF then begin new transaction
	if errors.Is(c.Error, ErrBreak) {
		return
	}
	if err = c.isEOF(); err != nil {
		c.Error = err
		return
	}
	c.tx = c.parentTx.Begin()
	return
}

func (c *Context[T]) Run(loopFn func(*Context[T])) *Context[T] {
	for !c.complete() {
		loopFn(c)
	}
	return c
}
