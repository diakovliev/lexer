package logger

// NopImpl is a logger that does nothing.
type NopImpl struct{}

// Nop returns a logger that does nothing.
func Nop() *NopImpl {
	return &NopImpl{}
}

// Error does nothing.
func (NopImpl) Error(format string, args ...any) {
}

// Info does nothing.
func (NopImpl) Info(format string, args ...any) {
}

// Warn does nothing.
func (NopImpl) Warn(format string, args ...any) {
}

// Debug does nothing.
func (NopImpl) Debug(format string, args ...any) {
}

// Trace does nothing.
func (NopImpl) Trace(format string, args ...any) {
}
