package message

type (
	// Factory is a message factory
	Factory[T any] interface {
		// Token creates a token message
		Token(level int, token T, value any, pos int, width int) Message[T]
		// Error creates an error message
		Error(level int, err error, value any, pos int, width int) Message[T]
	}

	DefaultFactoryImpl[T any] struct{}
)

func DefaultFactory[T any]() *DefaultFactoryImpl[T] {
	return &DefaultFactoryImpl[T]{}
}

func (f DefaultFactoryImpl[T]) Token(level int, token T, value any, pos int, width int) Message[T] {
	return Message[T]{
		Level: level,
		Type:  Token,
		Token: token,
		Value: value,
		Pos:   pos,
		Width: width,
	}
}

func (f DefaultFactoryImpl[T]) Error(level int, err error, value any, pos int, width int) Message[T] {
	return Message[T]{
		Level: level,
		Type:  Error,
		Value: &ErrorValue{
			Err:   err,
			Value: value,
		},
		Pos:   pos,
		Width: width,
	}
}
