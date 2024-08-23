package lexer

type (
	// StateFn represents a state function
	StateFn[T any] func(*Lexer[T]) StateFn[T]

	// State represents a state
	State[T any] interface {
		// State returns the next state function.
		State(*Lexer[T]) StateFn[T]
	}

	// StateFn2 represents a state function
	StateFn2[T any] func(*Lexer2[T]) StateFn2[T]

	// State represents a state
	State2[T any] interface {
		// State returns the next state function.
		State(*Lexer2[T]) StateFn2[T]
	}
)
