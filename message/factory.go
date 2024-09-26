package message

import "context"

type (
	// Factory is a message factory
	Factory[T any] interface {
		// Token creates a token message
		Token(ctx context.Context, level int, token T, value []byte, pos int, width int) (*Message[T], error)
		// Error creates an error message
		Error(ctx context.Context, level int, err error, buffer []byte, pos int, width int) (*Message[T], error)
	}

	// DefaultFactoryImpl is the default message factory implementation
	DefaultFactoryImpl[T any] struct{}
)

// DefaultFactory creates a default message factory instance.
func DefaultFactory[T any]() *DefaultFactoryImpl[T] {
	return &DefaultFactoryImpl[T]{}
}

// Token implements Factory interface.
func (f DefaultFactoryImpl[T]) Token(ctx context.Context, level int, token T, value []byte, pos int, width int) (msg *Message[T], err error) {
	msg = &Message[T]{}
	msg.Level = level
	msg.Type = Token
	msg.Token = token
	msg.Pos = pos
	msg.Width = width
	msg.Value = value
	return
}

// Error implements Factory interface.
func (f DefaultFactoryImpl[T]) Error(ctx context.Context, level int, userErr error, buffer []byte, pos int, width int) (msg *Message[T], err error) {
	msg = &Message[T]{}
	msg.Level = level
	msg.Type = Error
	msg.Pos = pos
	msg.Width = width
	errorValue := &ErrorValue{}
	errorValue.Err = userErr
	errorValue.Value = buffer
	msg.Value = errorValue
	return
}
