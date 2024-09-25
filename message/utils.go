package message

const preallocateCount = 1024

func growSlice[T any](slice []T, delta int) []T {
	newCap := cap(slice) + delta
	newSlice := make([]T, len(slice), newCap)
	copy(newSlice, slice)
	return newSlice
}

func preallocate[T any](n int) []T {
	return make([]T, n)
}
