package common

// Logger is the interface for logging.
type Logger interface {
	// Error logs a message using ERROR as log level.
	Error(format string, args ...any)
	// Info logs a message using INFO as log level.
	Info(format string, args ...any)
	// Warn logs a message using WARN as log level.
	Warn(format string, args ...any)
	// Debug logs a message using DEBUG as log level.
	Debug(format string, args ...any)
	// Trace logs a message using TRACE as log level.
	Trace(format string, args ...any)
	// Print logs a message.
	Print(format string, args ...any)
	// Fatal logs a message using FATAL as log level and panics.
	Fatal(format string, args ...any)
}
