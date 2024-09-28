package stack

// Stack implements a simple LIFO data structure.
type Stack[T any] []T

func New[T any](capacity int) Stack[T] {
	return make([]T, 0, capacity)
}

func (s Stack[T]) Push(v T) Stack[T] {
	return append(s, v)
}

func (s Stack[T]) Empty() bool {
	return len(s) == 0
}

func (s Stack[T]) Peek() T {
	l := len(s)
	if l == 0 {
		panic("stack is empty")
	}
	return s[l-1]
}

func (s Stack[T]) Pop() (Stack[T], T) {
	l := len(s)
	if l == 0 {
		panic("stack is empty")
	}
	return s[:l-1], s[l-1]
}
