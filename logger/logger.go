package logger

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Logger is a simple logger that writes to an io.Writer.
type Logger struct {
	sync.Mutex
	level    Level
	tmFormat string
	writer   io.Writer
}

// New creates a new logger with the given options.
func New(opts ...Option) (ret *Logger) {
	ret = &Logger{
		level:    Error,
		tmFormat: time.RFC3339Nano,
		writer:   io.Discard,
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (l *Logger) write(format string, args ...any) {
	l.Lock()
	defer l.Unlock()
	builder := &strings.Builder{}
	_, _ = fmt.Fprintf(builder, format, args...)
	_, _ = builder.WriteRune('\n')
	_, _ = l.writer.Write([]byte(builder.String()))
}

func (l *Logger) levelWrite(level Level, format string, args ...any) {
	if l.level < level {
		return
	}
	l.Lock()
	defer l.Unlock()
	builder := &strings.Builder{}
	_, _ = builder.WriteRune('[')
	_, _ = builder.WriteString(level.String())
	_, _ = builder.WriteString("] [")
	_, _ = builder.WriteString(time.Now().Format(l.tmFormat))
	_, _ = builder.WriteString("] ")
	// traceble levels need caller
	if level == Error || level == Fatal || level == Debug || level == Trace {
		stack := stacksPool.Get().(*stack)
		frame := stack.upFrame(4)
		fmt.Fprintf(builder, "[%s:%d] ", frame.Function, frame.Line)
		// fmt.Fprintf(builder, "%s:%d %s", frame.File, frame.Line, frame.Function)
		stacksPool.Put(stack)
	}
	_, _ = fmt.Fprintf(builder, format, args...)
	_, _ = builder.WriteRune('\n')
	_, _ = l.writer.Write([]byte(builder.String()))
}

// Error writes an error message to the log.
func (l *Logger) Error(format string, args ...any) {
	l.levelWrite(Error, format, args...)
}

// Info writes an info message to the log.
func (l *Logger) Info(format string, args ...any) {
	l.levelWrite(Info, format, args...)
}

// Warn writes a warning message to the log.
func (l *Logger) Warn(format string, args ...any) {
	l.levelWrite(Warn, format, args...)
}

// Debug writes a debug message to the log.
func (l *Logger) Debug(format string, args ...any) {
	l.levelWrite(Debug, format, args...)
}

// Trace writes a trace message to the log.
func (l *Logger) Trace(format string, args ...any) {
	l.levelWrite(Trace, format, args...)
}

// Print writes a message to the log.
func (l *Logger) Print(format string, args ...any) {
	l.write(format, args...)
}

// Fatal writes a fatal message to the log and panics.
func (l *Logger) Fatal(format string, args ...any) {
	l.levelWrite(Fatal, format, args...)
	panic(fmt.Errorf(format, args...))
}
