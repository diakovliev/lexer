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

	DefaultFactoryImpl[T any] struct{}
)

func DefaultFactory[T any]() *DefaultFactoryImpl[T] {
	return &DefaultFactoryImpl[T]{}
}

func (f DefaultFactoryImpl[T]) Token(ctx context.Context, level int, token T, value any, pos int, width int) (msg *Message[T], err error) {
	msg = &Message[T]{
		Level: level,
		Type:  Token,
		Token: token,
		Value: value,
		Pos:   pos,
		Width: width,
	}
	return
}

func (f DefaultFactoryImpl[T]) Error(ctx context.Context, level int, userErr error, value any, pos int, width int) (msg *Message[T], err error) {
	msg = &Message[T]{
		Level: level,
		Type:  Error,
		Value: &ErrorValue{
			Err:   userErr,
			Value: value,
		},
		Pos:   pos,
		Width: width,
	}
	return
}
