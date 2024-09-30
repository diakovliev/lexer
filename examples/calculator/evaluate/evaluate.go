package evaluate

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/diakovliev/lexer/examples/calculator/parser"
	"github.com/diakovliev/lexer/examples/calculator/vm"
)

var VM *vm.VM = vm.New()

var (
	// ErrVMError is returned when vm failed to execute the code.
	ErrVMError = errors.New("VM error")
	// ErrParserError is returned when parser failed to produce a code.
	ErrParserError = errors.New("parser error")
)

func Evaluate(input string) (result string, err error) {
	code, err := parser.New().Parse(bytes.NewBufferString(input))
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrParserError, err)
		return
	}
	if err = VM.PushCode(code).Run(); err != nil && !errors.Is(err, vm.ErrHalt) {
		err = fmt.Errorf("%w: %s", ErrVMError, err)
		return
	}
	value, err := VM.Peek()
	if err != nil {
		err = nil
		return
	}
	err = nil
	result = value.String()
	return
}
