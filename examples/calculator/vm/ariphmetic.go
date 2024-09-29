package vm

func add(vm *VM, op Cell, args ...Cell) (result *Cell, err error) {
	left := args[1]
	right := args[0]
	if left.IsWhole() && right.IsWhole() {
		result = &Cell{Op: Val, Value: left.AsInt64() + right.AsInt64()}
		return
	}
	result = &Cell{Op: Val, Value: left.AsFloat64() + right.AsFloat64()}
	return
}

func sub(vm *VM, op Cell, args ...Cell) (result *Cell, err error) {
	left := args[1]
	right := args[0]
	if left.IsWhole() && right.IsWhole() {
		result = &Cell{Op: Val, Value: left.AsInt64() - right.AsInt64()}
		return
	}
	result = &Cell{Op: Val, Value: left.AsFloat64() - right.AsFloat64()}
	return
}

func mul(vm *VM, op Cell, args ...Cell) (result *Cell, err error) {
	left := args[1]
	right := args[0]
	if left.IsWhole() && right.IsWhole() {
		result = &Cell{Op: Val, Value: left.AsInt64() * right.AsInt64()}
		return
	}
	result = &Cell{Op: Val, Value: left.AsFloat64() * right.AsFloat64()}
	return
}

func div(vm *VM, op Cell, args ...Cell) (result *Cell, err error) {
	left := args[1]
	right := args[0]
	if right.IsZero() {
		err = ErrDivByZero
		return
	}
	result = &Cell{Op: Val, Value: left.AsFloat64() / right.AsFloat64()}
	return
}
