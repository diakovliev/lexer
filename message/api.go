package message

type (
	Receiver[T any] func(Message[T]) error
)

func Dispose[T any](Message[T]) error {
	return nil
}
