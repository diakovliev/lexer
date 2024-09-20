package state

func AsSlice[T any](v ...T) (ret []T) {
	ret = append(ret, v...)
	return
}
