package lexer

// Option is a function that modifies the lexer's behavior.
type Option[T any] func(*Lexer[T])

// WithHistoryDepth sets the maximum number of items to keep in the history buffer.
func WithHistoryDepth[T any](depth int) Option[T] {
	return func(l *Lexer[T]) {
		l.historyDepth = depth
	}
}
