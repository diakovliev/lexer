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
	code, err := algo.Parse(algo.ShuntingYard(receiver.Slice))
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	vm := algo.NewVm(code)
	if err = vm.Execute(); err != nil && !errors.Is(err, algo.ErrVmComplete) {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = nil
	ret = fmt.Sprintf("%d", vm.Pop().Value)
	return
}

func main() {
	fmt.Print("The calculator example. Copyright (C) 2024, daemondzk@gmail.com.\n")
	fmt.Print("It supports whole nubers, brackets and basic ariphmetic operations: +, -, *, /.\n")
	fmt.Print("It is part of the github.com/diakovliev/lexer project.\n")
	fmt.Print("To exit press Ctrl+C.\n")
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
