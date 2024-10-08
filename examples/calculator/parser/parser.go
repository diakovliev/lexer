package parser

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/examples/calculator/number/parse"
	"github.com/diakovliev/lexer/examples/calculator/vm"
	"github.com/diakovliev/lexer/message"
)

var (
	ErrLexerError               = errors.New("lexer error")
	ErrUnknownToken             = errors.New("unknown token")
	ErrNotEnoughArgumentsForSet = errors.New("not enough arguments for 'set'")
	ErrNonNumberValueAllocation = errors.New("non number value allocation")
)

type (
	// Token is a lexer token.
	Token = *message.Message[grammar.Token]

	// Parse parses tokens into vm code
	Parser struct {
		code []vm.Cell
	}
)

var mapOp = map[grammar.Token]vm.OpCode{
	grammar.Comma:      vm.Comma,
	grammar.Bra:        vm.Bra,
	grammar.Ket:        vm.Ket,
	grammar.Plus:       vm.Add,
	grammar.Minus:      vm.Sub,
	grammar.Mul:        vm.Mul,
	grammar.Div:        vm.Div,
	grammar.Identifier: vm.Call,
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

func New() *Parser {
	return &Parser{
		code: make([]vm.Cell, 0),
	}
}

func (p *Parser) Receive(msgs []*message.Message[grammar.Token]) (err error) {
	for _, token := range msgs {
		if token.Type == message.Error {
			err = token.Value.(error)
			return
		}
		switch {
		case isNumber(token):
			number, parseErr := parse.ParseNumber(token.AsBytes())
			if parseErr != nil {
				err = parseErr
				return
			}
			p.code = append(p.code, vm.Cell{Op: vm.Val, Value: number})
			continue
		default:
			op, ok := mapOp[token.Token]
			if !ok {
				err = fmt.Errorf("%w: %d", ErrUnknownToken, token.Token)
				return
			}
			p.code = append(p.code, vm.Cell{Op: op, Value: token.AsString()})
		}
	}
	return
}

func (p *Parser) reset() {
	p.code = make([]vm.Cell, 0)
}

func (p *Parser) withoutCommas() []vm.Cell {
	var code []vm.Cell
	for _, cell := range p.code {
		if cell.Op != vm.Comma {
			code = append(code, cell)
		}
	}
	return code
}

func (p *Parser) Parse(input io.Reader) (code []vm.Cell, err error) {
	p.reset()
	lexer := grammar.New(input, p)
	if err = lexer.Run(context.Background()); !errors.Is(err, io.EOF) {
		err = fmt.Errorf("%w: %s", ErrLexerError, err)
		return
	}
	err = nil
	p.code = shuntingYard(p.code)
	code = p.withoutCommas()
	return
}
