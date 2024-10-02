package message

type (
	// Receiver is a type that can receive messages.
	Receiver[T any] interface {
		// Receive receives a message.
		Receive([]*Message[T]) error
	}

	// SliceReceiver is a receiver that stores messages in a slice.
	SliceReceiver[T any] struct {
		Slice []*Message[T]
	}

	// DisposeReceiver is a receiver that discards messages.
	DisposeReceiver[T any] struct{}
)

// Dispose returns a receiver that discards messages.
func Dispose[T any]() *DisposeReceiver[T] {
	return &DisposeReceiver[T]{}
}

// Receive implements the Receiver interface.
func (DisposeReceiver[T]) Receive([]*Message[T]) (err error) {
	return
}

// Slice returns a receiver that stores messages in a slice.
func Slice[T any]() *SliceReceiver[T] {
	return &SliceReceiver[T]{
		Slice: make([]*Message[T], 0, preallocateCount),
	}
}

// Receive implements the Receiver interface.
func (sr *SliceReceiver[T]) Receive(m []*Message[T]) (err error) {
	newLen := len(sr.Slice) + len(m)
	if newLen > cap(sr.Slice) {
		sr.Slice = growSlice(sr.Slice, newLen)
	}
	sr.Slice = append(sr.Slice, m...)
	return
}

// Reset resets the slice receiver to an empty state.
func (sr *SliceReceiver[T]) Reset() {
	sr.Slice = sr.Slice[:0]
}

// UserErrors returns all errors messages from the slice receiver.
func (sr *SliceReceiver[T]) UserErrors() (errs []*Message[T]) {
	errs = GetUserErrors[T](sr.Slice)
	return
}
