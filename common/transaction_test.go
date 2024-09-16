package common

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/diakovliev/lexer/logger"
	"github.com/stretchr/testify/assert"
)

func TestTransactionReader(t *testing.T) {

	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	tr := NewReader(logger, bytes.NewBufferString("this is test string"))

	// read first 4 bytes and rollback the transaction
	out := make([]byte, 4)
	rt := tr.Begin()
	n, err := rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("this"), out)
	assert.NoError(t, rt.Rollback())

	// read first 4 bytes and commit the transaction
	out = make([]byte, 4)
	rt = tr.Begin()
	n, err = rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte("this"), out)
	n, err = rt.Commit()
	assert.NoError(t, err)
	assert.Equal(t, 4, n)

	// read next 4 bytes and rollback the transaction
	out = make([]byte, 4)
	rt = tr.Begin()
	n, err = rt.Read(out)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, []byte(" is "), out)
	assert.NoError(t, rt.Rollback())
}

func TestTransactionReader2(t *testing.T) {

	logger := logger.New(
		logger.WithLevel(logger.Trace),
		logger.WithWriter(os.Stdout),
	)

	out := make([]byte, 4)

	tr := NewReader(logger, bytes.NewBufferString(""))
	assert.False(t, tr.eof)

	ctx0 := tr.Begin()

	cctx0 := ctx0.Begin()

	assert.False(t, cctx0.eof)
	_, err := cctx0.Read(out)
	assert.ErrorIs(t, err, io.EOF)
	assert.True(t, cctx0.eof)
	_, err = cctx0.Commit()
	assert.NoError(t, err)

	assert.True(t, ctx0.eof)
	_, err = ctx0.Read(out)
	assert.ErrorIs(t, err, io.EOF)
	assert.True(t, ctx0.eof)

	_, err = ctx0.Commit()
	assert.NoError(t, err)
	assert.True(t, tr.eof)
}
