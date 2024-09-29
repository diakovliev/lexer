package vm

func call(vm *VM, op Cell, args ...Cell) (result *Cell, err error) {
	result, err = vm.Call(op, args...)
	return
}
