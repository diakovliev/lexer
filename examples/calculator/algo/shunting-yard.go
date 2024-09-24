package algo

import (
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/message"
)

type (
	Token     = *message.Message[grammar.Token]
	Operators map[grammar.Token]Op
	Op        struct {
		Precedence int
		IsLeft     bool
	}
)

var Ops = Operators{
	grammar.Plus:  Op{Precedence: 0, IsLeft: true},
	grammar.Minus: Op{Precedence: 0, IsLeft: true},
	grammar.Mul:   Op{Precedence: 5, IsLeft: true},
	grammar.Div:   Op{Precedence: 5, IsLeft: true},
}

func (o Operators) Has(t Token) (ok bool) {
	_, ok = o[t.Token]
	return
}

func (o Operators) HasToken(t grammar.Token) (ok bool) {
	_, ok = o[t]
	return
}

func (o Operators) Precedence(t Token) (p int) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	p = op.Precedence
	return
}

func (o Operators) IsLeftAssociative(t Token) (ok bool) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	return op.IsLeft
}

func ShuntingYarg(tokens []Token) (output []Token) {
	var stack stack[Token]
	for _, token := range tokens {
		switch {
		case Ops.Has(token):
			for !stack.Empty() && Ops.Has(stack.Peek()) && Ops.Precedence(token) <= Ops.Precedence(stack.Peek()) {
				var token Token
				stack, token = stack.Pop()
				output = append(output, token)
			}
			stack = stack.Push(token)
		case token.Token == grammar.Bra:
			stack = stack.Push(token)
		case token.Token == grammar.Ket:
			for !stack.Empty() && stack.Peek().Token != grammar.Bra {
				var token Token
				stack, token = stack.Pop()
				output = append(output, token)
			}
			stack, _ = stack.Pop()
		default:
			output = append(output, token)
		}
	}
	for !stack.Empty() {
		var token Token
		stack, token = stack.Pop()
		output = append(output, token)
	}
	return
}
