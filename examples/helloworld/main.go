package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"unicode"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
)

type Token uint

const (
	Hello Token = iota
	World
)

func Lex(input string) (res []*message.Message[Token], err error) {
	receiver := message.Slice[Token]()
	lex := lexer.New(
		logger.New(),
		bytes.NewBufferString(input),
		message.DefaultFactory[Token](),
		receiver,
	).With(func(b state.Builder[Token]) []state.Update[Token] {
		return state.AsSlice[state.Update[Token]](
			b.Named("ignore spaces").
				RuneCheck(unicode.IsSpace).
				Repeat(state.CountBetween(1, math.MaxUint)).Omit(),
			b.Named("hello").String("hello").Emit(Hello),
			b.Named("world").String("world").Emit(World),
			b.Named("error").Rest().Error(errors.New("unexpected input")),
		)
	})
	err = lex.Run(context.Background())
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	// reset io.EOF
	err = nil
	res = receiver.Slice
	return
}

const (
	ps  = "? "
	out = "= "
)

func main() {

	fmt.Print("The helloworld example. Copyright (C) 2024, daemondzk@gmail.com.\n")
	fmt.Print("Licensed under the MIT license.\n")
	fmt.Print("Enter 'hello', 'world' or 'hello world' and press Enter.\n")
	fmt.Print("It is part of the 'github.com/diakovliev/lexer' project.\n")
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
		res, err := Lex(text)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			fmt.Printf("%s", ps)
			continue
		}
		for _, msg := range res {
			fmt.Printf("%s%+v\n", out, msg)
		}
		fmt.Print(ps)
	}
}
