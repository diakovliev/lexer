package common

// Ref is a reference to an interface.
type Ref[T any] struct {
	ref T
}

// NewRef creates a new Ref instance.
func NewRef[T any](inst T) Ref[T] {
	return Ref[T]{ref: inst}
}

// Deref returns underlying interface.
func (r Ref[T]) Deref() T {
	return r.ref
}
