package lexer

import (
	"bytes"
	"errors"
	"io"
)

type (
	AcceptorReceiver[T any] func(Message[T])

	Acceptor[T any] struct {
		Error    error
		buffer   *bytes.Buffer
		tx       *ReaderTransaction
		receiver AcceptorReceiver[T]
		delayed  []Message[T]
		done     bool
	}

	yielder[T any] struct {
		messages []Message[T]
	}
)

// see: ContextReceiver[T any]
func (y *yielder[T]) yield(m []Message[T]) {
	y.messages = append(y.messages, m...)
}

func NewAcceptor[T any](
	parentTx *ReaderTransaction,
	reveiver func(messages Message[T]),
) *Acceptor[T] {
	return &Acceptor[T]{
		tx:       parentTx.Begin(),
		receiver: reveiver,
		buffer:   bytes.NewBuffer(nil),
		delayed:  make([]Message[T], 0, 10),
	}
}

// internal receiver used when messages from the child
// acceptor are to be ignored.
func (ac *Acceptor[T]) ignore(Message[T]) {}

// has returns true if and only if internal buffer
// is not empty.
func (ac Acceptor[T]) has() (ret bool) {
	if ac.tx == nil {
		return
	}
	if ac.buffer == nil {
		return
	}
	ret = ac.buffer.Len() > 0
	return
}

// accept accepts or not next input rune from the input
// transaction using acceptFn predicate. If there is
// next rune, accept will return predicate result.
// If there is no more input, or any io error happend,
// accept will return error.
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

// complete clears underlying buffer and transaction
// pointer
func (ac *Acceptor[T]) complete() *Acceptor[T] {
	ac.buffer.Reset()
	ac.buffer = nil
	ac.tx = nil
	return ac
}

// isComplete checks is acceptor is in complete state.
func (ac Acceptor[T]) isComplete() bool {
	return ac.tx == nil
}

// resolve fixes position (commits) the underlying
// transaction and sets done flag to true, if there
// is no Error or Error is io.EOF. If Error is neither
// nil or io.EOF, done flag will be set to false.
func (ac *Acceptor[T]) resolve() *Acceptor[T] {
	if _, err := ac.tx.Commit(); err != nil {
		ac.Error = err
	}
	if ac.Error != nil {
		ac.done = errors.Is(ac.Error, io.EOF)
	} else {
		ac.done = true
	}
	return ac.complete()
}

// reject rejects underlying transaction.
func (ac *Acceptor[T]) reject() *Acceptor[T] {
	if err := ac.tx.Rollback(); err != nil {
		ac.Error = err
	}
	return ac.complete()
}

// Done returns value of the done flag.
// done flag indicates what the Acceptor is
// resolved with positive (true) or negative (false) result.
func (ac *Acceptor[T]) Done() bool {
	return ac.done
}

func (ac *Acceptor[T]) Fn(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	accepted, err := ac.accept(ac.tx, acceptFn)
	if err != nil || !accepted {
		ac.Error = err
		return ac.reject()
	}
	return ac
}

// Emit emits given message.
func (ac *Acceptor[T]) Emit(msgType MessageType, userType ...T) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	var msgUserType T
	switch {
	case len(userType) == 0:
	case len(userType) == 1:
		msgUserType = userType[0]
	case len(userType) > 1:
		panic("too many user types")
	}
	pos := int(ac.tx.Parent().Pos())
	value := make([]byte, ac.buffer.Len())
	width := copy(value, ac.buffer.Bytes())
	ac.receiver(Message[T]{
		Type:     msgType,
		UserType: msgUserType,
		Value:    value,
		Pos:      pos,
		Width:    width,
	})
	for _, msg := range ac.delayed {
		ac.receiver(msg)
	}
	ac.delayed = ac.delayed[:0]
	return ac.resolve()
}

// Emit2 is a shortcut for Emit(MessageType, T)
func (ac *Acceptor[T]) Emit2(userType T) *Acceptor[T] {
	return ac.Emit(User, userType)
}

func (ac *Acceptor[T]) Skip() *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	return ac.resolve()
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
			return ac.reject()
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
		return ac.reject()
	}
	return ac
}

func (ac *Acceptor[T]) While(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	return ac.acceptLoop(acceptFn)
}

func (ac *Acceptor[T]) Until(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	return ac.acceptLoop(func(r rune) bool { return !acceptFn(r) })
}

func (ac *Acceptor[T]) Count(count int) *Acceptor[T] {
	accepted := 0
	return ac.While(func(r rune) (ret bool) {
		ret = accepted < count
		accepted++
		return
	})
}

func (ac *Acceptor[T]) String(match string) *Acceptor[T] {
	buffer := bytes.NewBuffer([]byte(match))
	return ac.While(func(r rune) (ret bool) {
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
			ac.reject()
		}
		return
	})
}

func (ac *Acceptor[T]) AnyString(matches ...string) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	for _, match := range matches {
		if acceptor := NewAcceptor(ac.tx, ac.ignore).String(match); acceptor.has() {
			// get data from matchContext and commit child transaction
			ac.buffer.Write(acceptor.buffer.Bytes())
			_, ac.Error = acceptor.tx.Commit()
			return ac
		}
	}
	return ac.reject()
}

func (ac *Acceptor[T]) AnyFn(acceptFns ...func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	for _, accept := range acceptFns {
		if acceptor := NewAcceptor(ac.tx, ac.ignore).Fn(accept); acceptor.has() {
			// get data from acceptContext and commit child transaction
			ac.buffer.Write(acceptor.buffer.Bytes())
			_, ac.Error = acceptor.tx.Commit()
			return ac
		}
	}
	return ac.reject()
}

func (ac *Acceptor[T]) OptionallyWhile(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).While(acceptFn); acceptor.has() {
		// get data from matchContext and commit child transaction
		ac.buffer.Write(acceptor.buffer.Bytes())
		_, ac.Error = acceptor.tx.Commit()
	}
	return ac
}

func (ac *Acceptor[T]) OptionallyUntil(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).Until(acceptFn); acceptor.has() {
		// get data from matchContext and commit child transaction
		ac.buffer.Write(acceptor.buffer.Bytes())
		_, ac.Error = acceptor.tx.Commit()
	}
	return ac
}

func (ac *Acceptor[T]) FollowedByFn(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).Fn(acceptFn); acceptor.has() {
		ac.Error = acceptor.tx.Rollback()
		return ac
	}
	return ac.reject()
}

func (ac *Acceptor[T]) OptionallyFollowedByFn(acceptFn func(rune) bool) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	if acceptor := NewAcceptor(ac.tx, ac.ignore).Fn(acceptFn); acceptor.has() {
		ac.Error = acceptor.tx.Rollback()
	}
	return ac
}

func (ac *Acceptor[T]) If(ctxFn func(*Context[T])) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	yield := &yielder[T]{}
	if childContext := NewContext(ac.tx, yield.yield).Run(ctxFn); childContext.Error != nil && !errors.Is(childContext.Error, ErrBreak) {
		ac.Error = childContext.Error
		return ac.reject()
	}
	if len(yield.messages) == 0 {
		return ac.reject()
	}
	ac.delayed = append(ac.delayed, yield.messages...)
	return ac
}

func (ac *Acceptor[T]) Optionally(ctxFn func(*Context[T])) *Acceptor[T] {
	if ac.isComplete() {
		return ac
	}
	yield := &yielder[T]{}
	if childContext := NewContext(ac.tx, yield.yield).Run(ctxFn); childContext.Error != nil && !errors.Is(childContext.Error, ErrBreak) {
		ac.Error = childContext.Error
		return ac.reject()
	}
	ac.delayed = append(ac.delayed, yield.messages...)
	return ac
}
