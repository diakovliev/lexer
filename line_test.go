package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLine(t *testing.T) {
	ln := &line{}
	assert.Equal(t, 0, ln.number())
	assert.Equal(t, 0, ln.start())
	ln = ln.next(10)
	assert.Equal(t, 1, ln.number())
	assert.Equal(t, 10, ln.start())
}
