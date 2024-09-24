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
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
)

const (
	ps = ">> "
)

func evaluate(input string) (ret string, err error) {
	fmt.Printf("input: '%s'\n", input)
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stderr),
	)
	factory := message.DefaultFactory[grammar.Token]()
	receiver := message.Slice[grammar.Token]()
	lexer := lexer.New(logger, bytes.NewBufferString(input), factory, receiver).With(grammar.BuildState)
	lexErr := lexer.Run(context.TODO())
	for _, msg := range receiver.Slice {
		fmt.Printf("> %s\n", msg)
	}
	if !errors.Is(lexErr, io.EOF) {
		err = lexErr
		return
	}

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
		fmt.Printf("res: %s\n", res)
		fmt.Print(ps)
	}
}
