package state

type (
	// RunePredicate is a function that takes rune and returns true if it should be accepted.
	RunePredicate func(rune) bool

	// BytePredicate is a function that takes byte and returns true if it should be accepted.
	BytePredicate func(byte) bool
)

// IsByte returns a function that checks if the given byte is equal to the sample.
func IsByte(sample byte) func(byte) bool {
	return func(in byte) bool {
		return sample == in
	}
}

// IsRune returns a function that checks if the given rune is equal to the sample.
func IsRune(sample rune) func(rune) bool {
	return func(in rune) bool {
		return sample == in
	}
}

// True returns a function that always returns true.
func True[T any]() func(T) bool {
	return func(_ T) bool { return true }
}

// False returns a function that always returns false.
func False[T any]() func(T) bool {
	return func(_ T) bool { return false }
}

// Not returns a function that negates the given predicate.
func Not[T any](pred func(T) bool) (negative func(T) bool) {
	return func(v T) bool { return !pred(v) }
}

// Or returns a function that checks if at least one of the given predicates is true.
func Or[T any](predicates ...func(T) bool) func(T) bool {
	return func(t T) bool {
		for _, pred := range predicates {
			if pred(t) {
				return true
			}
		}
		return false
	}
}

// And returns a function that checks if all of the given predicates are true.
func And[T any](predicates ...func(T) bool) func(T) bool {
	return func(t T) bool {
		for _, pred := range predicates {
			if !pred(t) {
				return false
			}
		}
		return true
	}
}
