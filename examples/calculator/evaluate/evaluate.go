package evaluate

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/examples/calculator/parser"
	"github.com/diakovliev/lexer/examples/calculator/vm"
)

var VM *vm.VM

var (
	// ErrVMError is returned when vm failed to execute the code.
	ErrVMError = errors.New("VM error")
	// ErrParserError is returned when parser failed to produce a code.
	ErrParserError = errors.New("parser error")
)

func Init(opts ...vm.Option) {
	VM = vm.New(opts...)
}

func Evaluate(input string) (err error) {
	if VM == nil {
		panic("vm is not initialized")
	}
	code, err := parser.New().Parse(bytes.NewBufferString(input))
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrParserError, err)
		return
	}
	if err = VM.PushCode(code).Run(); errors.Is(err, vm.ErrHalt) {
		err = nil
	}
	VM.Output("= ", "S")
	return
}
