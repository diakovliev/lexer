package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexeme(t *testing.T) {
	lm := &lexeme{}
	assert.Equal(t, 0, lm.start())
	assert.Equal(t, 0, lm.pos())
	assert.Equal(t, 0, lm.width())
	lm.add(10)
	assert.Equal(t, 0, lm.start())
	assert.Equal(t, 10, lm.width())
	assert.Equal(t, 10, lm.pos())

	lm = lm.next()
	assert.Equal(t, 10, lm.start())
	assert.Equal(t, 10, lm.pos())
	assert.Equal(t, 0, lm.width())
	lm.add(20)
	assert.Equal(t, 10, lm.start())
	assert.Equal(t, 20, lm.width())
	assert.Equal(t, 30, lm.pos())

	lm = lm.from(50)
	assert.Equal(t, 50, lm.start())
	assert.Equal(t, 50, lm.pos())
	assert.Equal(t, 0, lm.width())
	lm.add(20)
	assert.Equal(t, 50, lm.start())
	assert.Equal(t, 20, lm.width())
	assert.Equal(t, 70, lm.pos())

	undo := lm.restore()
	lm.add(30)
	assert.Equal(t, 50, lm.start())
	assert.Equal(t, 50, lm.width())
	assert.Equal(t, 100, lm.pos())

	undo()
	assert.Equal(t, 50, lm.start())
	assert.Equal(t, 20, lm.width())
	assert.Equal(t, 70, lm.pos())

	lm.reset()
	assert.Equal(t, 50, lm.start())
	assert.Equal(t, 50, lm.pos())
	assert.Equal(t, 0, lm.width())
}
