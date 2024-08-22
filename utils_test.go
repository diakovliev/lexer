package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRune(t *testing.T) {
	assert.True(t, Rune('a')('a'))
	assert.False(t, Rune('a')('b'))
}

func TestRunes(t *testing.T) {
	assert.True(t, Runes("ab")('a'))
	assert.True(t, Runes("ab")('b'))
	assert.False(t, Runes("ab")('c'))
	assert.False(t, Runes("ab")('d'))
}
