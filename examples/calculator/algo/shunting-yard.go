package algo

import (
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/message"
)

type (
	// Token is a lexer token.
	Token = *message.Message[grammar.Token]

	// Operators is a map of operators and their properties.
	Operators map[grammar.Token]Op

	// Op is a operator properties.
	Op struct {
		Precedence int
		IsLeft     bool
	}
)

// Ops is a map of operators and their properties.
var Ops = Operators{
	grammar.Plus:  Op{Precedence: 0, IsLeft: true},
	grammar.Minus: Op{Precedence: 0, IsLeft: true},
	grammar.Mul:   Op{Precedence: 5, IsLeft: true},
	grammar.Div:   Op{Precedence: 5, IsLeft: true},
}

// Has checks if the token is an operator.
func (o Operators) Has(t Token) (ok bool) {
	_, ok = o[t.Token]
	return
}

// Has checks if the token is an operator.
func (o Operators) HasToken(t grammar.Token) (ok bool) {
	_, ok = o[t]
	return
}

// Precedence returns the precedence of an operator.
func (o Operators) Precedence(t Token) (p int) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	p = op.Precedence
	return
}

// IsLeftAssociative checks if the operator is left associative.
func (o Operators) IsLeftAssociative(t Token) (ok bool) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	return op.IsLeft
}

// ShuntingYard implements the Shunting-yard algorithm for converting an infix expression to a postfix one.
func ShuntingYard(tokens []Token) (output []Token) {
	stack := makeStack[Token](100)
	output = make([]Token, 0, len(tokens))
	for _, token := range tokens {
		switch {
		case Ops.Has(token):
			var precedence int
			// Check if the operator is right-associative and adjust the precedence comparison accordingly
			if Ops.IsLeftAssociative(token) {
				precedence = Ops.Precedence(token)
			} else {
				precedence = Ops.Precedence(token) - 1
			}
			// Check if the stack is empty or if the top of the stack has a lower precedence than the current operator
			for !stack.Empty() && Ops.Has(stack.Peek()) && precedence <= Ops.Precedence(stack.Peek()) {
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
