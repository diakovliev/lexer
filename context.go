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

func (c *Context[T]) Accept(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).Accept(acceptFn)
	return c.Current
}

func (c *Context[T]) AcceptUntil(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptUntil(acceptFn)
	return c.Current
}

func (c *Context[T]) AcceptWhile(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptWhile(acceptFn)
	return c.Current
}

func (c *Context[T]) AcceptCount(count int) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptCount(count)
	return c.Current
}

func (c *Context[T]) AcceptString(match string) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptString(match)
	return c.Current
}

func (c *Context[T]) AcceptAnyStringFrom(matches ...string) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptAnyStringFrom(matches...)
	return c.Current
}

func (c *Context[T]) AcceptAnyFrom(acceptFns ...func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptAnyFrom(acceptFns...)
	return c.Current
}

func (c *Context[T]) OptionallyAcceptWhile(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).OptionallyAcceptWhile(acceptFn)
	return c.Current
}

func (c *Context[T]) OptionallyAcceptUntil(acceptFn func(rune) bool) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).OptionallyAcceptUntil(acceptFn)
	return c.Current
}

func (c *Context[T]) AcceptContext(ctxFn func(*Context[T])) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).AcceptContext(ctxFn)
	return c.Current
}

func (c *Context[T]) OptionallyAcceptContext(ctxFn func(*Context[T])) *Acceptor[T] {
	c.Current = NewAcceptor(c.tx, c.addMessage).OptionallyAcceptContext(ctxFn)
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
