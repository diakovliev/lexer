package lexer

type Option[T any] func(*Lexer[T])

func WithHistoryDepth[T any](depth int) Option[T] {
	return func(l *Lexer[T]) {
		l.historyDepth = depth
	}
}
