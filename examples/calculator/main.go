package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/examples/calculator/algo"
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
)

const (
	ps = ">> "
)

func evaluate(input string) (ret string, err error) {
	receiver := message.Slice[grammar.Token]()

	lexer := lexer.New(
		logger.New(
			logger.WithLevel(logger.Trace),
			logger.WithWriter(os.Stderr),
		),
		bytes.NewBufferString(input),
		message.DefaultFactory[grammar.Token](),
		receiver).
		With(grammar.BuildState(true))

	lexErr := lexer.Run(context.TODO())
	if !errors.Is(lexErr, io.EOF) {
		err = lexErr
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	// for _, msg := range receiver.Slice {
	// 	fmt.Printf("> %s\n", msg)
	// }
	bpf := algo.ShuntingYarg(receiver.Slice)
	// for _, msg := range bpf {
	// 	fmt.Printf("bpf> %s\n", msg)
	// }
	vmData, err := algo.Parse(bpf)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	vm := algo.NewVm()
	err = vm.Execute(vmData)
	if err != nil && !errors.Is(err, algo.ErrVmComplete) {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = nil
	ret = fmt.Sprintf("%d", vm.Pop().Value)
	return
}

func main() {
	fmt.Print(ps)
	for {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			continue
		}
		if scanner.Err() != nil {
			fmt.Printf("ERROR: %s\n", scanner.Err())
			return
		}
		text := scanner.Text()
		if len(strings.TrimSpace(text)) == 0 {
			fmt.Printf("%s", ps)
			continue
		}
		res, err := evaluate(text)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			fmt.Printf("%s", ps)
			continue
		}
		fmt.Printf("%s\n", res)
		fmt.Print(ps)
	}
}
