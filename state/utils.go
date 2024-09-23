package state

func AsSlice[T any](v ...T) (ret []T) {
	ret = append(ret, v...)
	return
}

func byteEqual(sample byte) func(byte) bool {
	return func(in byte) bool {
		return sample == in
	}
}

func runeEqual(sample rune) func(rune) bool {
	return func(in rune) bool {
		return sample == in
	}
}

func alwaysTrue[T any]() func(T) bool {
	return func(_ T) bool { return true }
}

// func alwaysFalse[T any]() func(T) bool {
// 	return func(_ T) bool { return false }
// }

func negatePredicate[T any](positive func(T) bool) (negative func(T) bool) {
	return func(v T) bool { return !positive(v) }
}
