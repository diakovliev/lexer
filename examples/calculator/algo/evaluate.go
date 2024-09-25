package algo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
)

var (
	// ErrVMError is returned when vm failed to execute the code.
	ErrVMError = errors.New("VM error")
	// ErrNoResult is returned when vm where stopped normally but there were no result on top of the stack.
	ErrNoResult = errors.New("no result")
	// ErrLexerError is returned when lexer failed to produce a tokens stream.
	ErrLexerError = errors.New("lexer error")
	// ErrParserError is returned when parser failed to produce a code.
	ErrParserError = errors.New("parser error")
)

func Evaluate(input string) (result string, err error) {
	receiver := message.Slice[grammar.Token]()
	lexer := lexer.New(
		logger.New(
			logger.WithLevel(logger.Trace),
			logger.WithWriter(os.Stderr),
		),
		bytes.NewBufferString(input),
		message.DefaultFactory[grammar.Token](),
		receiver,
		lexer.WithHistoryDepth[grammar.Token](1),
	).With(grammar.New(true))
	if err = lexer.Run(context.TODO()); !errors.Is(err, io.EOF) {
		err = fmt.Errorf("%w: %s", ErrLexerError, err)
		return
	}
	code, err := Parse(ShuntingYard(receiver.Slice))
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrParserError, err)
		return
	}
	vm := NewVM(code)
	if err = vm.Run(); err != nil && !errors.Is(err, ErrVMHalt) {
		err = fmt.Errorf("%w: %s", ErrVMError, err)
		return
	}
	value, err := vm.Pop()
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrNoResult, err)
		return
	}
	err = nil
	result = fmt.Sprintf("%d", value.Value)
	return
}
