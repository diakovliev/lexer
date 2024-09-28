package vm

import "math"

// Cell is a token in virtual machine
type Cell struct {
	Op    OpCode
	Value any
}

func (vc Cell) IsZero() bool {
	if vc.Op != Val {
		panic("not a value")
	}
	return vc.AsInt64() == 0 && vc.AsFloat64() == 0
}

func (vc Cell) IsWhole() bool {
	if vc.Op != Val {
		panic("not a value")
	}
	_, ok := vc.Value.(int64)
	return ok
}

func (vc Cell) AsInt64() (i int64) {
	if vc.Op != Val {
		panic("not a value")
	}
	i, ok := vc.Value.(int64)
	if !ok {
		f, ok := vc.Value.(float64)
		if !ok {
			panic("not a number")
		}
		i = int64(math.Round(f))
	}
	return
}

func (vc Cell) AsFloat64() (f float64) {
	if vc.Op != Val {
		panic("not a value")
	}
	f, ok := vc.Value.(float64)
	if !ok {
		i, ok := vc.Value.(int64)
		if !ok {
			panic("not a number")
		}
		f = float64(i)
	}
	return
}
