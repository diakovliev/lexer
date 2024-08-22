package lexer

type (
	// StateFn represents a state function
	StateFn[T any] func(*Lexer[T]) StateFn[T]

	// State represents a state
	State[T any] interface {
		// State returns the next state function.
		State(*Lexer[T]) StateFn[T]
	}
)
