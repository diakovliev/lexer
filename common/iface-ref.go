package common

// IfaceRef is a reference to an interface.
type IfaceRef[T any] struct {
	Ref T
}

// Ref creates a new IfaceRef instance.
func Ref[T any](inst T) IfaceRef[T] {
	return IfaceRef[T]{Ref: inst}
}
