package logger_test

import (
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
)

func TestLogger(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	logger.Info("test message %s", "123")
	logger.Warn("test")

	logger.Error("test")
	logger.Debug("test")
	logger.Trace("test")
}
