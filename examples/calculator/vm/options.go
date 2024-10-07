package vm

import "io"

// Option is a function that takes a pointer to VM and modifies it in some way.
type Option func(*VM)

// WithOutput sets the output writer for the VM. This is where any print statements will be written to. If not set, it defaults to os.Stdout.
func WithOutput(output io.Writer) Option {
	return func(vm *VM) {
		vm.output = output
	}
}
