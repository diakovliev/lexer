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

	r := New(logger, bytes.NewBufferString(testString))

	assert.True(t, r.Has())
	n, err := r.Fetch(-1)
	assert.NoError(t, err)
	assert.Zero(t, n)
	n, err = r.Fetch(0)
	assert.NoError(t, err)
	assert.Zero(t, n)

	for pos := startPos; pos < len(testString); pos += 1 {
		end := pos + bufferSize
		if end > len(testStringData) {
			end = len(testStringData)
		}
		expected := string(testStringData[pos:end])
		t.Run(
			fmt.Sprintf("1st pass: pos: %d buffer size: %d expected: %s", pos, bufferSize, expected),
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

	buffer, _, err := r.Buffer()
	assert.NoError(t, err)
	assert.True(t, len(buffer) > 0)

	for pos := startPos; pos < len(testString); pos += 1 {
		end := pos + bufferSize
		if end > len(testStringData) {
			end = len(testStringData)
		}
		expected := string(testStringData[pos:end])
		t.Run(
			fmt.Sprintf("2nd pass: pos: %d buffer size: %d expected: %s", pos, bufferSize, expected),
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

	r := New(logger, bytes.NewBufferString(testString))

	for i := 0; i < txCount; i++ {
		pos := i * bufferSize
		end := pos + bufferSize
		if end > len(testStringData) {
			end = len(testStringData)
		}
		expected := string(testStringData[pos:end])
		t.Run(
			fmt.Sprintf("tx %d, buffer size: %d, expected: %s", i, bufferSize, expected),
			func(t *testing.T) {
				state := r.Begin().Deref()
				buffer := make([]byte, bufferSize)
				n, err := state.Read(buffer)
				if n < bufferSize {
					assert.ErrorIs(t, err, io.EOF)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, testStringData[pos:pos+n], buffer[:n])
				err = AsTx(state).Commit()
				assert.NoError(t, err)
			},
		)
	}
}
