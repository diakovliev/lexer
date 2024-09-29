package algo

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/examples/calculator/vm"
	"github.com/diakovliev/lexer/message"
)

var (
	ErrUnknownToken             = errors.New("unknown token")
	ErrNotEnoughArgumentsForSet = errors.New("not enough argiments for 'set'")
	ErrNonNumberValueAllocation = errors.New("non number value allocation")
)

const setFunctionName = "set"

type (
	// Token is a lexer token.
	Token = *message.Message[grammar.Token]
)

var numberBases = map[string]int{
	grammar.BinNumberPrefixes[0]: 2,
	grammar.BinNumberPrefixes[1]: 2,
	grammar.OctNumberPrefixes[0]: 8,
	grammar.OctNumberPrefixes[1]: 8,
	grammar.HexNumberPrefixes[0]: 16,
	grammar.HexNumberPrefixes[1]: 16,
}

// parse float in arbitraty base
// Exponents chart:
// Pos:       ...  |0   1   2   3   4    5    6   | ...
// Digits:    ...  |N   N   N   .   N    N    N   | ...
// Exponents: ...  |b^2 b^1 b^0     b^-1 b^-2 b^-3| ...
func parseFloat(buffer []byte, base int) (result float64, err error) {
	// I hope we have enough precision)
	dotPos := bytes.IndexByte(buffer, '.')
	maxExponent := dotPos - 1
	startBase := math.Pow(float64(base), float64(maxExponent))
	for i := 0; i < len(buffer); i++ {
		if i == dotPos {
			continue
		}
		delta := float64(buffer[i]-'0') * startBase
		if math.IsNaN(delta) {
			// no sense to continue
			break
		}
		result += delta
		startBase /= float64(base)
	}
	return
}

func parseNumber(buffer []byte) (any, error) {
	isNegative := bytes.HasPrefix(buffer, []byte("-"))
	if isNegative {
		buffer = buffer[1:]
	}
	if bytes.HasPrefix(buffer, []byte("+")) {
		buffer = buffer[1:]
	}
	base := 10
	for prefix, pBase := range numberBases {
		if bytes.Contains(buffer, []byte(prefix)) {
			base = pBase
			buffer = buffer[len(prefix):]
			break
		}
	}
	if !bytes.ContainsFunc(buffer, grammar.IsNumberDot) {
		var result int64
		// whole
		result, err := strconv.ParseInt(string(buffer), base, 64)
		if err != nil {
			return nil, err
		}
		if isNegative {
			result = -result
		}
		return result, nil
	}
	// float
	var result float64
	result, err := parseFloat(buffer, base)
	if err != nil {
		return nil, err
	}
	if isNegative {
		result = -result
	}
	return result, nil
}

var mapOp = map[grammar.Token]vm.OpCode{
	grammar.Plus:  vm.Add,
	grammar.Minus: vm.Sub,
	grammar.Mul:   vm.Mul,
	grammar.Div:   vm.Div,
}

func isNumber(token Token) bool {
	switch token.Token {
	case grammar.DecFraction,
		grammar.BinFraction,
		grammar.OctFraction,
		grammar.HexFraction,
		grammar.DecNumber,
		grammar.BinNumber,
		grammar.OctNumber,
		grammar.HexNumber:
		return true
	default:
		return false
	}
}

func filterOutCommas(tokens []Token) (result []Token) {
	for _, token := range tokens {
		if token.Token != grammar.Comma {
			result = append(result, token)
		}
	}
	return
}

// Parse parses tokens into vm code
func Parse(tokens []Token) (data []vm.Cell, err error) {
	tokens = filterOutCommas(tokens)
	data = make([]vm.Cell, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		token := tokens[i]
		if token.Type == message.Error {
			err = token.Value.(error)
			return
		}
		switch {
		case isNumber(token):
			number, parseErr := parseNumber(token.AsBytes())
			if parseErr != nil {
				err = parseErr
				return
			}
			data = append(data, vm.Cell{Op: vm.Val, Value: number})
			continue
		case token.Token == grammar.Identifier:
			// CALL identifier
			data = append(data, []vm.Cell{
				{Op: vm.Ident, Value: token.AsString()},
				{Op: vm.Call, Value: nil},
			}...)
			continue
		default:
			op, ok := mapOp[token.Token]
			if !ok {
				err = fmt.Errorf("%w: %d", ErrUnknownToken, token.Token)
				return
			}
			data = append(data, vm.Cell{Op: op})
		}
	}
	return
}
