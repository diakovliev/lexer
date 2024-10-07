package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/diakovliev/lexer/examples/calculator/evaluate"
	"github.com/diakovliev/lexer/examples/calculator/vm"
)

const (
	ps  = "? "
	out = "= "
)

//go:embed welcome.txt
var welcome string

func init() {
	evaluate.Init(vm.WithOutput(os.Stdout))
}

func main() {
	fmt.Print(welcome)
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
		if err := evaluate.Evaluate(text); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			fmt.Printf("%s", ps)
			continue
		}
		fmt.Print(ps)
	}
}
