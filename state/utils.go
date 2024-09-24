package state

func AsSlice[T any](v ...T) (ret []T) {
	ret = append(ret, v...)
	return
}

func IsByte(sample byte) func(byte) bool {
	return func(in byte) bool {
		return sample == in
	}
}

func IsRune(sample rune) func(rune) bool {
	return func(in rune) bool {
		return sample == in
	}
}

func True[T any]() func(T) bool {
	return func(_ T) bool { return true }
}

func False[T any]() func(T) bool {
	return func(_ T) bool { return false }
}

func Not[T any](pred func(T) bool) (negative func(T) bool) {
	return func(v T) bool { return !pred(v) }
}

func Or[T any](preds ...func(T) bool) func(T) bool {
	return func(t T) bool {
		for _, pred := range preds {
			if pred(t) {
				return true
			}
		}
		return false
	}
}

func And[T any](preds ...func(T) bool) func(T) bool {
	return func(t T) bool {
		for _, pred := range preds {
			if !pred(t) {
				return false
			}
		}
		return true
	}
}
