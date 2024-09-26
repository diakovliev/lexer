package message

const preallocateCount = 1024

func growSlice[T any](slice []T, delta int) []T {
	newCap := cap(slice) + delta
	newSlice := make([]T, len(slice), newCap)
	copy(newSlice, slice)
	return newSlice
}

// GetUserErrors returns all errors messages from the messages slice.
func GetUserErrors[T any](slice []*Message[T]) (errs []*Message[T]) {
	for _, m := range slice {
		if m.Type == Error {
			errs = append(errs, m)
		}
	}
	return
}
