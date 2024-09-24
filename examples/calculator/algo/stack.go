package algo

type stack[T any] []T

func (s stack[T]) Push(v T) stack[T] {
	return append(s, v)
}

func (s stack[T]) Empty() bool {
	return len(s) == 0
}

func (s stack[T]) Peek() T {
	l := len(s)
	if l == 0 {
		panic("stack is empty")
	}
	return s[l-1]
}

func (s stack[T]) Pop() (stack[T], T) {
	l := len(s)
	if l == 0 {
		panic("stack is empty")
	}
	return s[:l-1], s[l-1]
}
