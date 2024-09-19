package logger

import "io"

// Option is a function that takes a Logger and modifies its configuration.
type Option func(*Logger)

// WithLevel sets the level of logs to be written. If not set, it will default to Error.
func WithLevel(level Level) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// WithWriter sets the writer to write logs to. If not set, it will default to io.Dispose.
func WithWriter(writer io.Writer) Option {
	return func(l *Logger) {
		l.writer = writer
	}
}

// WithTmFormat sets the time format for log messages.
func WithTmFormat(tmFormat string) Option {
	return func(l *Logger) {
		l.tmFormat = tmFormat
	}
}
