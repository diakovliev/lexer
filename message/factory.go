package message

import "context"

type (
	// Factory is a message factory
	Factory[T any] interface {
		// Token creates a token message
		Token(ctx context.Context, level int, token T, value any, pos int, width int) (*Message[T], error)
		// Error creates an error message
		Error(ctx context.Context, level int, err error, value any, pos int, width int) (*Message[T], error)
	}

	DefaultFactoryImpl[T any] struct {
		// preallocatedMessages []Message[T]
		// preallocatedErrors   []ErrorValue
	}
)

// DefaultFactory creates a default message factory instance.
func DefaultFactory[T any]() *DefaultFactoryImpl[T] {
	return &DefaultFactoryImpl[T]{
		// preallocatedMessages: preallocate[Message[T]](preallocateCount),
		// preallocatedErrors:   preallocate[ErrorValue](preallocateCount),
	}
}

func (f *DefaultFactoryImpl[T]) getPreallocatedMessage() *Message[T] {
	msg := Message[T]{}
	return &msg
	// if len(f.preallocatedMessages) == 0 {
	// 	f.preallocatedMessages = preallocate[Message[T]](preallocateCount)
	// }
	// msg := f.preallocatedMessages[len(f.preallocatedMessages)-1]
	// f.preallocatedMessages = f.preallocatedMessages[:len(f.preallocatedMessages)-1]
	// return &msg
}

func (f *DefaultFactoryImpl[T]) getPreallocatedError() *ErrorValue {
	errorValue := ErrorValue{}
	return &errorValue
	// if len(f.preallocatedErrors) == 0 {
	// 	f.preallocatedErrors = preallocate[ErrorValue](preallocateCount)
	// }
	// errorValue := f.preallocatedErrors[len(f.preallocatedErrors)-1]
	// f.preallocatedErrors = f.preallocatedErrors[:len(f.preallocatedErrors)-1]
	// return &errorValue
}

// Token implements Factory interface.
func (f DefaultFactoryImpl[T]) Token(ctx context.Context, level int, token T, value any, pos int, width int) (msg *Message[T], err error) {
	msg = f.getPreallocatedMessage()
	msg.Level = level
	msg.Type = Token
	msg.Token = token
	msg.Pos = pos
	msg.Width = width
	msg.Value = value
	return
}

// Error implements Factory interface.
func (f DefaultFactoryImpl[T]) Error(ctx context.Context, level int, userErr error, value any, pos int, width int) (msg *Message[T], err error) {
	msg = f.getPreallocatedMessage()
	msg.Level = level
	msg.Type = Error
	msg.Pos = pos
	msg.Width = width
	errorValue := f.getPreallocatedError()
	errorValue.Err = userErr
	errorValue.Value = value
	msg.Value = errorValue
	return
}
