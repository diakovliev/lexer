//go:build js && wasm
// +build js,wasm

package main

import (
	"bytes"
	_ "embed"
	"io"
	"strings"
	"syscall/js"

	"github.com/diakovliev/lexer/examples/calculator/evaluate"
	"github.com/diakovliev/lexer/examples/calculator/vm"
)

var output bytes.Buffer

//go:embed welcome.txt
var welcome string

func evaluateWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		builder := strings.Builder{}
		if len(args) != 1 {
			builder.WriteString("ERROR: Invalid no of arguments passed\n")
			return builder.String()
		}
		if err := evaluate.Evaluate(args[0].String()); err != nil {
			builder.WriteString("ERROR: " + err.Error() + "\n")
		}
		ret, err := io.ReadAll(&output)
		if err != nil {
			builder.WriteString("ERROR: " + err.Error() + "\n")
			return builder.String()
		}
		output.Truncate(0)
		builder.Write(ret)
		return builder.String()
	})
}

func welcomeWrapper() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		return welcome
	})
}

func init() {
	evaluate.Init(vm.WithOutput(&output))
}

func main() {
	js.Global().Set("wasmWelcome", welcomeWrapper())
	js.Global().Set("wasmEvaluate", evaluateWrapper())
	<-make(chan struct{})
}
