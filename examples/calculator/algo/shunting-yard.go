package algo

import (
	"github.com/diakovliev/lexer/examples/calculator/grammar"
	"github.com/diakovliev/lexer/examples/calculator/stack"
)

type (
	// syOperators is a map of operators and their properties.
	syOperators map[grammar.Token]syOp

	// syOp is a operator properties.
	syOp struct {
		Precedence int
		IsLeft     bool
	}
)

// syOps is a map of operators and their properties.
var syOps = syOperators{
	grammar.Plus:  syOp{Precedence: 0, IsLeft: true},
	grammar.Minus: syOp{Precedence: 0, IsLeft: true},
	grammar.Mul:   syOp{Precedence: 5, IsLeft: true},
	grammar.Div:   syOp{Precedence: 5, IsLeft: true},
}

// Has checks if the token is an operator.
func (o syOperators) Has(t Token) (ok bool) {
	_, ok = o[t.Token]
	return
}

// Has checks if the token is an operator.
func (o syOperators) HasToken(t grammar.Token) (ok bool) {
	_, ok = o[t]
	return
}

// Precedence returns the precedence of an operator.
func (o syOperators) Precedence(t Token) (p int) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	p = op.Precedence
	return
}

// IsLeftAssociative checks if the operator is left associative.
func (o syOperators) IsLeftAssociative(t Token) (ok bool) {
	op, ok := o[t.Token]
	if !ok {
		panic("unreachable")
	}
	return op.IsLeft
}

// ShuntingYard implements the Shunting-yard algorithm for converting an infix expression to a postfix one.
func ShuntingYard(tokens []Token) (output []Token) {
	stack := stack.New[Token](100)
	output = make([]Token, 0, len(tokens))
	for _, token := range tokens {
		switch {
		case syOps.Has(token):
			var precedence int
			// Check if the operator is right-associative and adjust the precedence comparison accordingly
			if syOps.IsLeftAssociative(token) {
				precedence = syOps.Precedence(token)
			} else {
				precedence = syOps.Precedence(token) - 1
			}
			// Check if the stack is empty or if the top of the stack has a lower precedence than the current operator
			for !stack.Empty() && syOps.Has(stack.Peek()) && precedence <= syOps.Precedence(stack.Peek()) {
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
