package logger_test

import (
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	logger.Print("test")
	logger.Info("test message %s", "123")
	logger.Warn("test")

	logger.Error("test")
	logger.Debug("test")
	logger.Trace("test")
	// logger.Fatal("test")
	assert.Panics(t, func() { logger.Fatal("test") })
}
