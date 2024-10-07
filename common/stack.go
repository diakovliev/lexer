package common

import "errors"

var (
	// ErrEmptyStack indicates that a stack operation was attempted on an empty stack.
	ErrEmptyStack = errors.New("empty stack")
)

// Stack implements a simple LIFO data structure.
type Stack[T any] struct {
	arr []T
}

// NewStack returns a new stack with the given capacity.
func NewStack[T any](capacity int) *Stack[T] {
	return &Stack[T]{
		arr: make([]T, 0, capacity),
	}
}

// Empty returns true if the stack is empty.
func (s Stack[T]) Empty() bool {
	return len(s.arr) == 0
}

// Push pushes a value onto the stack.
func (s *Stack[T]) Push(v T) *Stack[T] {
	s.arr = append(s.arr, v)
	return s
}

// Peek returns the top value on the stack without removing it. If the stack is empty, an error is returned.
func (s Stack[T]) Peek() (v T, err error) {
	l := len(s.arr)
	if l == 0 {
		err = ErrEmptyStack
		return
	}
	v = s.arr[l-1]
	return
}

// Pop removes and returns the top value from the stack. If the stack is empty, an error is returned.
func (s *Stack[T]) Pop() (v T, err error) {
	l := len(s.arr)
	if l == 0 {
		err = ErrEmptyStack
		return
	}
	v = s.arr[l-1]
	s.arr = s.arr[:l-1]
	return
}

// AsSlice returns the stack as a slice. The returned slice is not a copy, so modifying it will modify the stack.
func (s Stack[T]) AsSlice() []T {
	return s.arr[:]
}
