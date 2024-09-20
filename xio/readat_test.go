package xio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/stretchr/testify/assert"
)

func TestReader(t *testing.T) {
	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	testString := "1234567890"
	testStringData := []byte(testString)
	startPos := 1
	bufferSize := 3

	r := NewReadAt(logger, bytes.NewBufferString(testString))

	for pos := startPos; pos < len(testString); pos += 1 {
		t.Run(
			fmt.Sprintf("2nd pass: pos: %d buffer size: %d expected: %s", pos, bufferSize, string(testStringData[pos:pos+bufferSize])),
			func(t *testing.T) {
				buffer := make([]byte, bufferSize)
				n, err := r.ReadAt(int64(pos), buffer)
				if n < bufferSize {
					assert.ErrorIs(t, err, io.EOF)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testStringData[pos:pos+n], buffer[:n])
			},
		)
	}

	for pos := startPos; pos < len(testString); pos += 1 {
		t.Run(
			fmt.Sprintf("2nd pass: pos: %d buffer size: %d expected: %s", pos, bufferSize, string(testStringData[pos:pos+bufferSize])),
			func(t *testing.T) {
				buffer := make([]byte, bufferSize)
				n, err := r.ReadAt(int64(pos), buffer)
				if n < bufferSize {
					assert.ErrorIs(t, err, io.EOF)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testStringData[pos:pos+n], buffer[:n])
			},
		)
	}
}

func TestReader_Truncate(t *testing.T) {

	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	testString := "1234567890"
	testStringData := []byte(testString)
	txCount := 4
	bufferSize := 3

	r := NewReadAt(logger, bytes.NewBufferString(testString))

	for i := 0; i < txCount; i++ {
		pos := i * bufferSize
		t.Run(
			fmt.Sprintf("tx %d, buffer size: %d, expected: %s", i, bufferSize, string(testStringData[pos:pos+bufferSize])),
			func(t *testing.T) {
				tx := r.Begin()
				buffer := make([]byte, bufferSize)
				n, err := tx.Read(buffer)
				if n < bufferSize {
					assert.ErrorIs(t, err, io.EOF)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testStringData[pos:pos+n], buffer[:n])
				err = tx.Commit()
				assert.NoError(t, err)
				// assert.NoError(t, r.Truncate(int64(pos)))
			})
	}
}
