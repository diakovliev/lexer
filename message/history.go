package message

type (
	// History is a message history
	History[T any] interface {
		Receiver[T]
		// Get returns the history. The latest element is at the end of the list.
		Get() []*Message[T]
	}

	// RememberImpl is an implementation of History.
	// It remembers the last messages and returns them
	RememberImpl[T any] struct {
		receiver  Receiver[T]
		messages  []*Message[T]
		keepCount int
	}

	// ForgetImpl is an implementation of History.
	// It forgets the last messages.
	ForgetImpl[T any] struct {
		receiver Receiver[T]
	}
)

// Remember returns a message history.
// It remembers the last messages and returns them.
// keepCount is the number of last messages to remember.
func Remember[T any](receiver Receiver[T], keepCount int) *RememberImpl[T] {
	return &RememberImpl[T]{
		receiver:  receiver,
		messages:  make([]*Message[T], 0, keepCount),
		keepCount: keepCount,
	}
}

// Receive implements the Receiver interface
func (h *RememberImpl[T]) Receive(m *Message[T]) (err error) {
	err = h.receiver.Receive(m)
	if err != nil {
		return
	}
	h.messages = append(h.messages, m)
	if len(h.messages) > h.keepCount {
		h.messages = h.messages[len(h.messages)-h.keepCount:]
	}
	return
}

// Get implements the History interface
func (h *RememberImpl[T]) Get() []*Message[T] {
	return h.messages
}

// Forget returns a message history.
// It forgets the last messages.
func Forget[T any](receiver Receiver[T]) *ForgetImpl[T] {
	return &ForgetImpl[T]{
		receiver: receiver,
	}
}

// Receive implements the Receiver interface
func (h *ForgetImpl[T]) Receive(m *Message[T]) (err error) {
	return h.receiver.Receive(m)
}

// Get implements the History interface
func (h *ForgetImpl[T]) Get() []*Message[T] {
	return nil
}
