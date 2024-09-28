package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/diakovliev/lexer/examples/calculator/algo"
)

const (
	ps  = "? "
	out = "= "
)

func main() {
	fmt.Print("The calculator example. Copyright (C) 2024, daemondzk@gmail.com.\n")
	fmt.Print("Licensed under the MIT license.\n")
	fmt.Print("It supports numbers, brackets and basic arithmetic operations: +, -, *, /.\n")
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
		res, err := algo.Evaluate(text)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			fmt.Printf("%s", ps)
			continue
		}
		fmt.Printf("%s%s\n", out, res)
		fmt.Print(ps)
	}
}
