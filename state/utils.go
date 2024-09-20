package state

func AsSlice[T any](v ...T) (ret []T) {
	ret = append(ret, v...)
	return
}

func AsState[T any](v ...*Chain[T]) (ret []State[T]) {
	for _, chain := range v {
		ret = append(ret, chain)
	}
	return
}
