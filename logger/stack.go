package logger

import (
	"runtime"
	"sync"
)

const maxStackSize = 100

type stack struct {
	pc []uintptr
}

func newStack() *stack {
	return &stack{
		pc: make([]uintptr, maxStackSize),
	}
}

func (s *stack) upFrame(upFrame int) (frame runtime.Frame) {
	n := runtime.Callers(upFrame, s.pc)
	frames := runtime.CallersFrames(s.pc[:n])
	frame, _ = frames.Next()
	return
}

var stacksPool = sync.Pool{
	New: func() any { return newStack() },
}
