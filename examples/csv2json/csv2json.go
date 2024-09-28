package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/diakovliev/lexer/examples/csv2json/csv"
)

var (
	inputFile  string
	outputFile string
	separator  string
	withHeader bool
)

func init() {
	flag.StringVar(&inputFile, "i", "", "Input file, if no - stdin will be used.")
	flag.StringVar(&outputFile, "o", "", "Output file, if no - stdout will be used.")
	flag.StringVar(&separator, "s", ",", "Separator, ',' by default.")
	flag.BoolVar(&withHeader, "wh", false, "Treat the first line as a header with column names.")
}

func main() {
	flag.Parse()
	input := os.Stdin
	var err error
	if inputFile != "" {
		inputFile, err := os.OpenFile(inputFile, os.O_RDONLY, 0o644)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
		defer inputFile.Close()
		input = inputFile
	}
	output := os.Stdout
	if outputFile != "" {
		outputFile, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}
		defer outputFile.Close()
		output = outputFile
	}
	err = csv.Do(input, output, separator[0], withHeader)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
}
