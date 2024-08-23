package lexer

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransactionReader(t *testing.T) {

	tr := NewTransactionReader(bytes.NewBufferString("this is test string"))

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
