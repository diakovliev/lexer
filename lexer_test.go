package lexer

import (
	"errors"
	"regexp"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

type (
	testMessageType int
	testStateError  struct{}
	testStateEOF    struct{}
	runer           struct {
		s []byte
	}
)

const (
	NonUser testMessageType = iota
	Identifier
)

func (r runer) Get(i int) (ret rune) {
	idx := 0
	width := 0
	for {
		dRet, w := utf8.DecodeRune(r.s[width:])
		if dRet == utf8.RuneError {
			panic("rune error")
		}
		ret = dRet
		width += w
		idx++
		if idx > i {
			break
		}
	}
	return
}

func (tse *testStateError) State(lex *Lexer[testMessageType]) StateFn[testMessageType] {
	return lex.Break("test error")
}

func (tseof *testStateEOF) State(lex *Lexer[testMessageType]) StateFn[testMessageType] {
	for !lex.IsEOF() {
		switch {
		case lex.Skip(unicode.IsSpace):
			continue
		case lex.AcceptString("hello") || lex.AcceptString("world"):
			_ = lex.Emit(Identifier)
			continue
		}
	}
	_ = lex.EOF()
	return nil
}

func TestError(t *testing.T) {
	lex := New([]byte("hello world"), &testStateError{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())
	lex.Do()
	assert.Error(t, lex.Error)
	assert.True(t, errors.Is(lex.Error, ErrLexerError))
	assert.Nil(t, lex.Last())
}

func TestEOF(t *testing.T) {
	lex := New([]byte("hello world"), &testStateEOF{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())
	lex.Do()
	assert.NoError(t, lex.Error)
	assert.True(t, lex.IsEOF())
	assert.Nil(t, lex.Last())
}

func TestNext(t *testing.T) {
	input := []byte("hello world")
	runer := &runer{s: input}
	lex := New(input, &testStateEOF{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())
	for i := 0; i < len(input); i++ {
		expectedRune := runer.Get(i)
		r, undo, err := lex.Next()
		assert.NoError(t, err)
		assert.Equal(t, expectedRune, r)
		undo()
		r, _, err = lex.Next()
		assert.NoError(t, err)
		assert.Equal(t, expectedRune, r)
	}
	assert.True(t, lex.IsEOF())
	_, _, err := lex.Next()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errEOF))
	assert.True(t, lex.IsEOF())
}

func TestNextInvalidRune(t *testing.T) {
	input := []byte("\xf0\x28\x8c\x28")
	lex := New(input, &testStateEOF{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())
	_, _, err := lex.Next()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrLexerError))
	assert.True(t, lex.IsEOF())
}

func TestRegexp(t *testing.T) {

	input := []byte("hello world")

	lex := New(input, &testStateEOF{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())

	assert.False(t, lex.PeekRegexp(regexp.MustCompile(`garbage0`)))
	assert.False(t, lex.AcceptRegexp(regexp.MustCompile(`garbage1`)))
	assert.False(t, lex.SkipRegexp(regexp.MustCompile(`garbage2`)))

	assert.True(t, lex.PeekRegexp(regexp.MustCompile(`hello`)))
	assert.True(t, lex.AcceptRegexp(regexp.MustCompile(`hello`)))
	assert.Equal(t, "hello", lex.String())
	assert.NoError(t, lex.Emit(Identifier))

	assert.True(t, lex.PeekRegexp(regexp.MustCompile(`\s`)))
	assert.True(t, lex.SkipRegexp(regexp.MustCompile(`\s`)))

	assert.True(t, lex.PeekRegexp(regexp.MustCompile(`world`)))
	assert.True(t, lex.AcceptRegexp(regexp.MustCompile(`world`)))
	assert.Equal(t, "world", lex.String())
	assert.NoError(t, lex.Emit(Identifier))
}

func TestSkip(t *testing.T) {

	input := []byte("   hello \tworld")

	lex := New(input, &testStateEOF{})
	assert.NotNil(t, lex)
	assert.False(t, lex.IsEOF())

	assert.True(t, lex.SkipWhile(unicode.IsSpace))
	assert.True(t, lex.AcceptString("hello"))
	assert.Equal(t, "hello", lex.String())
	assert.NoError(t, lex.Emit(Identifier))
	assert.False(t, lex.AcceptString("world"))
	assert.True(t, lex.SkipAnyFrom(" \t\n\r"))
	assert.True(t, lex.Skip(Rune('\t')))
	assert.True(t, lex.AcceptString("world"))
	assert.Equal(t, "world", lex.String())
	assert.NoError(t, lex.Emit(Identifier))

	assert.True(t, lex.IsEOF())
}
