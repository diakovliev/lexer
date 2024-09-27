package state

// AsSlice returns a slice of the given values.
func AsSlice[T any](v ...T) (ret []T) {
	ret = append(ret, v...)
	return
}
