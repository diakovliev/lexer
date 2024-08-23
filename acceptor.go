package lexer

import (
	"bytes"
	"errors"
	"io"
)

type Acceptor[T any] struct {
	Error   error
	buffer  *bytes.Buffer
	tx      *ReaderTransaction
	emit    func(message Message[T])
	delayed []Message[T]
	done    bool
}

func NewAcceptor[T any](
	parentTx *ReaderTransaction,
	emit func(messages Message[T]),
) *Acceptor[T] {
	return &Acceptor[T]{
		tx:      parentTx.Begin(),
		emit:    emit,
		buffer:  bytes.NewBuffer(nil),
		delayed: make([]Message[T], 0, 10),
	}
}

func (ac *Acceptor[T]) ignore(Message[T]) {}

func (ac Acceptor[T]) getData() (pos int, data []byte, width int) {
	pos = int(ac.tx.Parent().Pos())
	data = make([]byte, ac.buffer.Len())
	copy(data, ac.buffer.Bytes())
	width = ac.buffer.Len()
	return
}

func (ac Acceptor[T]) Has() (ret bool) {
	if ac.tx == nil {
		return
	}
	if ac.buffer == nil {
		return
	}
	ret = ac.buffer.Len() > 0
	return
}

func (ac Acceptor[T]) accept(tx *ReaderTransaction, acceptFn func(rune) bool) (accepted bool, err error) {
	data, r, err := NextRuneFrom(tx)
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	if errors.Is(err, io.EOF) {
		if len(data) < 1 {
			return
		} else {
			err = nil
		}
	}
	if accepted = acceptFn(r); accepted {
		_, err = ac.buffer.Write(data)
	}
	return
}

func (ac *Acceptor[T]) complete(resolve bool) *Acceptor[T] {
	if resolve {
		if _, err := ac.tx.Commit(); err != nil {
			ac.Error = err
		}
		if ac.Error != nil {
			ac.done = errors.Is(ac.Error, io.EOF)
		} else {
			ac.done = true
		}
	} else {
		if err := ac.tx.Rollback(); err != nil {
			ac.Error = err
		}
	}
	ac.buffer.Reset()
	ac.buffer = nil
	ac.tx = nil
	return ac
}

func (ac *Acceptor[T]) Done() bool {
	return ac.done
}

func (ac *Acceptor[T]) Accept(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	accepted, err := ac.accept(ac.tx, acceptFn)
	if err != nil || !accepted {
		ac.Error = err
		return ac.complete(false)
	}
	return ac
}

func (ac *Acceptor[T]) Emit(msgType MessageType, userType T) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	pos, value, width := ac.getData()
	ac.emit(Message[T]{
		Type:     msgType,
		UserType: userType,
		Value:    value,
		Pos:      pos,
		Width:    width,
	})
	for _, msg := range ac.delayed {
		ac.emit(msg)
	}
	ac.delayed = ac.delayed[:0]
	return ac.complete(true)
}

func (ac *Acceptor[T]) Drop() *Acceptor[T] {
	var msg T
	return ac.Emit(Drop, msg)
}

func (ac *Acceptor[T]) Skip() *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	return ac.complete(true)
}

func (ac *Acceptor[T]) acceptLoop(
	acceptFn func(rune) bool,
) *Acceptor[T] {
	loopTx := ac.tx.Begin()
	count := 0
	for {
		accepted, err := ac.accept(loopTx, acceptFn)
		// break loop if context is resolved
		if ac.tx == nil || ac.Error != nil {
			return ac
		}
		if err != nil && !errors.Is(err, io.EOF) {
			ac.Error = err
			return ac.complete(false)
		}
		if errors.Is(err, io.EOF) {
			ac.Error = err
			if _, err = loopTx.Commit(); err != nil {
				ac.Error = err
				return ac
			}
			break
		}
		if !accepted {
			_, ac.Error = loopTx.Unread().Commit()
			break
		}
		count++
	}
	if count == 0 {
		return ac.complete(false)
	}
	return ac
}

func (ac *Acceptor[T]) AcceptWhile(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	return ac.acceptLoop(acceptFn)
}

func (ac *Acceptor[T]) AcceptUntil(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	return ac.acceptLoop(func(r rune) bool { return !acceptFn(r) })
}

func (ac *Acceptor[T]) AcceptCount(count int) *Acceptor[T] {
	accepted := 0
	return ac.AcceptWhile(func(r rune) (ret bool) {
		ret = accepted < count
		accepted++
		return
	})
}

func (ac *Acceptor[T]) AcceptString(match string) *Acceptor[T] {
	buffer := bytes.NewBuffer([]byte(match))
	return ac.AcceptWhile(func(r rune) (ret bool) {
		if buffer.Len() == 0 {
			return
		}
		_, br, err := NextRuneFrom(buffer)
		if err != nil {
			return
		}
		ret = r == br
		if !ret {
			// resolve context to break loop
			// see acceptLoop
			ac.complete(false)
		}
		return
	})
}

func (ac *Acceptor[T]) AcceptAnyStringFrom(matches ...string) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	for _, match := range matches {
		if acceptor := NewAcceptor(ac.tx, ac.ignore).AcceptString(match); acceptor.Has() {
			// get data from matchContext and commit child transaction
			ac.buffer.Write(acceptor.buffer.Bytes())
			_, ac.Error = acceptor.tx.Commit()
			return ac
		}
	}
	return ac.complete(false)
}

func (ac *Acceptor[T]) AcceptAnyFrom(acceptFns ...func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	for _, accept := range acceptFns {
		if acceptor := NewAcceptor(ac.tx, ac.ignore).Accept(accept); acceptor.Has() {
			// get data from acceptContext and commit child transaction
			ac.buffer.Write(acceptor.buffer.Bytes())
			_, ac.Error = acceptor.tx.Commit()
			return ac
		}
	}
	return ac.complete(false)
}

func (ac *Acceptor[T]) OptionallyAcceptWhile(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).AcceptWhile(acceptFn); acceptor.Has() {
		// get data from matchContext and commit child transaction
		ac.buffer.Write(acceptor.buffer.Bytes())
		_, ac.Error = acceptor.tx.Commit()
	}
	return ac
}

func (ac *Acceptor[T]) OptionallyAcceptUntil(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).AcceptUntil(acceptFn); acceptor.Has() {
		// get data from matchContext and commit child transaction
		ac.buffer.Write(acceptor.buffer.Bytes())
		_, ac.Error = acceptor.tx.Commit()
	}
	return ac
}

func (ac *Acceptor[T]) FollowedBy(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).Accept(acceptFn); acceptor.Has() {
		ac.Error = acceptor.tx.Rollback()
		return ac
	}
	return ac.complete(false)
}

func (ac *Acceptor[T]) OptionallyFollowedBy(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).Accept(acceptFn); acceptor.Has() {
		ac.Error = acceptor.tx.Rollback()
	}
	return ac
}

type yielder[T any] struct {
	messages []Message[T]
}

func (y *yielder[T]) yield(m []Message[T]) {
	y.messages = append(y.messages, m...)
}

func (ac *Acceptor[T]) AcceptContext(ctxFn func(*Context[T])) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	yield := &yielder[T]{}
	if childContext := NewContext(ac.tx, yield.yield).Run(ctxFn); childContext.Error != nil && !errors.Is(childContext.Error, ErrBreak) {
		ac.Error = childContext.Error
		return ac.complete(false)
	}
	if len(yield.messages) == 0 {
		return ac.complete(false)
	}
	ac.delayed = append(ac.delayed, yield.messages...)
	return ac
}

func (ac *Acceptor[T]) OptionallyAcceptContext(ctxFn func(*Context[T])) *Acceptor[T] {
	if ac.tx == nil {
		return ac
	}
	yield := &yielder[T]{}
	if childContext := NewContext(ac.tx, yield.yield).Run(ctxFn); childContext.Error != nil && !errors.Is(childContext.Error, ErrBreak) {
		ac.Error = childContext.Error
		return ac.complete(false)
	}
	ac.delayed = append(ac.delayed, yield.messages...)
	return ac
}
