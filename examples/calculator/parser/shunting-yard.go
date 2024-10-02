package parser

import (
	"github.com/diakovliev/lexer/examples/calculator/stack"
	"github.com/diakovliev/lexer/examples/calculator/vm"
)

type (
	// syOperators is a map of operators and their properties.
	syOperators map[vm.OpCode]syOp

	// syOp is a operator properties.
	syOp struct {
		Precedence int
		IsLeft     bool
	}
)

// syOps is a map of operators and their properties.
var syOps = syOperators{
	vm.Comma: syOp{Precedence: 0, IsLeft: true},
	vm.Add:   syOp{Precedence: 0, IsLeft: true},
	vm.Sub:   syOp{Precedence: 0, IsLeft: true},
	vm.Mul:   syOp{Precedence: 5, IsLeft: true},
	vm.Div:   syOp{Precedence: 5, IsLeft: true},
	vm.Call:  syOp{Precedence: 10, IsLeft: true},
}

// has checks if the token is an operator.
func (o syOperators) has(c vm.Cell) (ok bool) {
	_, ok = o[c.Op]
	return
}

// Has checks if the token is an operator.
func (o syOperators) hasToken(c vm.Cell) (ok bool) {
	_, ok = o[c.Op]
	return
}

// precedence returns the precedence of an operator.
func (o syOperators) precedence(c vm.Cell) (p int) {
	op, ok := o[c.Op]
	if !ok {
		panic("unreachable")
	}
	p = op.Precedence
	return
}

// isLeftAssociative checks if the operator is left associative.
func (o syOperators) isLeftAssociative(c vm.Cell) (ok bool) {
	op, ok := o[c.Op]
	if !ok {
		panic("unreachable")
	}
	return op.IsLeft
}

// shuntingYard implements the Shunting-yard algorithm for converting an infix expression to a postfix one.
func shuntingYard(input []vm.Cell) (code []vm.Cell) {
	stack := stack.New[vm.Cell](100)
	code = make([]vm.Cell, 0, len(input))
	for _, cell := range input {
		switch {
		case syOps.has(cell):
			var precedence int
			// Check if the operator is right-associative and adjust the precedence comparison accordingly
			if syOps.isLeftAssociative(cell) {
				precedence = syOps.precedence(cell)
			} else {
				precedence = syOps.precedence(cell) - 1
			}
			// Check if the stack is empty or if the top of the stack has a lower precedence than the current operator
			for !stack.Empty() && syOps.has(stack.Peek()) && precedence <= syOps.precedence(stack.Peek()) {
				var cell vm.Cell
				stack, cell = stack.Pop()
				code = append(code, cell)
			}
			stack = stack.Push(cell)
		case cell.Op == vm.Bra:
			stack = stack.Push(cell)
		case cell.Op == vm.Ket:
			for !stack.Empty() && stack.Peek().Op != vm.Bra {
				var cell vm.Cell
				stack, cell = stack.Pop()
				code = append(code, cell)
			}
			stack, _ = stack.Pop()
		default:
			code = append(code, cell)
		}
	}
	for !stack.Empty() {
		var cell vm.Cell
		stack, cell = stack.Pop()
		code = append(code, cell)
	}
	return
}
