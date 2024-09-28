package xio

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/stretchr/testify/assert"
)

func TestTx(t *testing.T) {

	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	r := New(logger, bytes.NewBufferString("this is test string"))

	// read first 4 bytes and rollback the transaction
	out := make([]byte, 4)
	rt := r.Begin().Deref()
	n, err := rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("this"), out)
	assert.NoError(t, rt.(*state).Rollback())

	// read first 4 bytes and commit the transaction
	out = make([]byte, 4)
	rt = r.Begin().Deref()
	n, err = rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("this"), out)
	err = rt.(*state).Commit()
	assert.NoError(t, err)

	// read next 4 bytes and rollback the transaction
	out = make([]byte, 4)
	rt = r.Begin().Deref()
	n, err = rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte(" is "), out)
	assert.NoError(t, rt.(*state).Rollback())
}

func TestTx2(t *testing.T) {

	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	out := make([]byte, 4)

	r := New(logger, bytes.NewBufferString(""))

	tx0 := r.Begin().Deref()

	ctx0 := tx0.(*state).Begin().Deref()

	_, err := ctx0.(*state).Read(out)
	assert.ErrorIs(t, err, io.EOF)
	err = ctx0.(*state).Commit()
	assert.NoError(t, err)

	_, err = tx0.(*state).Read(out)
	assert.ErrorIs(t, err, io.EOF)

	err = tx0.(*state).Commit()
	assert.NoError(t, err)
}
