package lexer

import (
	"errors"
	"fmt"
	"io"
)

var errSkip = errors.New("skip")

type Lexer2[T any] struct {
	// Error is the lexer error
	Error error
	// lexer start state
	startState State2[T]
	// callback is called when a message is
	callback func(Message[T]) error
	// current line
	// line *line
	// current lexeme
	// lexeme *lexeme2
	// reader is the input source of data.
	reader *TransactionReader
	// tx is a current reader tx
	tx *ReaderTransaction
}

func New2[T any](reader io.Reader, startState State2[T]) (ret *Lexer2[T]) {
	ret = &Lexer2[T]{
		startState: startState,
		reader:     NewTransactionReader(reader),
	}
	ret.tx = ret.reader.Begin()
	return
}

// func (lex *Lexer2[T]) Next() (ret rune, undo func(), err error) {
// 	if lex.Error != nil {
// 		err = lex.Error
// 		return
// 	}
// 	undo = lex.lexeme.restore()
// 	// if lex.IsEOF() {
// 	// 	err = errEOF
// 	// 	return
// 	// }

// 	ret, w := utf8.DecodeRune(lex.input[lex.lexeme.pos():])
// 	if ret == utf8.RuneError {
// 		err = fmt.Errorf("%w: invalid rune", ErrLexerError)
// 		lex.Error = err
// 		return
// 	}
// 	lex.lexeme.add(w)
// 	return
// }

func (lex *Lexer2[T]) Next(tx *ReaderTransaction) (data []byte, r rune, err error) {
	data, r, err = NextRuneFrom(tx)
	if err != nil && !errors.Is(err, io.EOF) {
		lex.Error = err
		return
	}
	if errors.Is(err, io.EOF) {
		lex.Error = errEOF
		if len(data) > 0 {
			err = nil
		}
	}
	return
}

func (lex *Lexer2[T]) emit(msgType MessageType, userType T) (err error) {
	// var value []byte
	// if msgType != NL && msgType != EOF {
	// 	value = lex.Lexeme()
	// }
	// message := Message[T]{
	// 	Type:     msgType,
	// 	UserType: userType,
	// 	Value:    value,
	// }
	// if err = lex.call(message); err != nil {
	// 	return
	// }
	// lex.remember(message)
	// if msgType == NL {
	// 	lex.line = lex.line.next(lex.lexeme.pos())
	// }
	// lex.lexeme = lex.lexeme.next()
	return
}

// Emit emits a token of the specified kind and calls the callback function if it is set.
//
// The kind parameter specifies the kind of token to emit.
// The function returns an error if there was an issue calling the callback.
// It returns nil otherwise.
func (lex *Lexer2[T]) Emit(msgType T) (err error) {
	return lex.emit(User, msgType)
}

// NL is a function that is used to emit line character lexeme in the lexer.
// It takes no parameters and returns an error.
func (lex *Lexer2[T]) NL() (err error) {
	var userType T
	return lex.emit(NL, userType)
}

// EOF is a function that is used to emit EOF lexeme in the lexer.
// It takes no parameters and returns an error.
func (lex *Lexer2[T]) EOF() (err error) {
	var userType T
	return lex.emit(EOF, userType)
}

func (lex *Lexer2[T]) Accept(acceptFn func(rune) bool) (accepted bool) {
	if lex.Error != nil {
		return
	}
	_, r, err := lex.Next(lex.tx)
	if err != nil {
		return
	}
	if accepted = acceptFn(r); accepted {
		// if len(subStates) > 0 {
		// 	for _, subState := range subStates {
		// 		if subErr := subState(lex); subErr != nil && !errors.Is(subErr, errSkip) {
		// 			err = subErr
		// 			break
		// 		}
		// 	}
		// 	if err != nil {

		// 	}
		// }
		if _, err := lex.tx.Commit(); err != nil {
			lex.Error = err
			return
		}
		//lex.lexeme.add(data, r)
	} else if err := lex.tx.Rollback(); err != nil {
		lex.Error = err
	}
	lex.tx = lex.reader.Begin()
	return
}

func (lex *Lexer2[T]) WithCallback(callback func(Message[T]) error) *Lexer2[T] {
	lex.callback = callback
	return lex
}

func (lex *Lexer2[T]) IsEOF() bool {
	// TODO: implement
	return false
}

// Do executes the Lexer's state machine.
//
// It iterates through each state function and calls it until a nil state function is returned.
func (lex *Lexer2[T]) Do() *Lexer2[T] {
	for stateFn := lex.startState.State; stateFn != nil; {
		stateFn = stateFn(lex)
	}
	if lex.Error != nil {
		return lex
	} else if !lex.IsEOF() {
		lex.Error = fmt.Errorf("%w: incomplete lexical analysis", ErrLexerError)
	}
	return lex
}
