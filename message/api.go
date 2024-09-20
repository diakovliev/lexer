package message

type (
	Receiver[T any] interface {
		Receive(Message[T]) error
	}

	SliceReceiver[T any] struct {
		Slice []Message[T]
	}

	DisposeReceiver[T any] struct{}
)

func Dispose[T any]() *DisposeReceiver[T] {
	return &DisposeReceiver[T]{}
}

func (DisposeReceiver[T]) Receive(Message[T]) (err error) {
	return
}

func Slice[T any]() *SliceReceiver[T] {
	return &SliceReceiver[T]{}
}

func (sr *SliceReceiver[T]) Receive(m Message[T]) (err error) {
	sr.Slice = append(sr.Slice, m)
	return
}

func (sr *SliceReceiver[T]) Reset() {
	sr.Slice = sr.Slice[:0]
}

func (sr *SliceReceiver[T]) EmitTo(r Receiver[T]) (err error) {
	for _, m := range sr.Slice {
		if err = r.Receive(m); err != nil {
			return err
		}
	}
	sr.Reset()
	return
}
