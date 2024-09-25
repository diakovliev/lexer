package grammar

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/diakovliev/lexer/message"
	"github.com/stretchr/testify/assert"
)

func TestGrammar(t *testing.T) {
	type testCase struct {
		name      string
		input     string
		wantError error
		want      []message.Message[Token]
	}
	tests := []testCase{
		// TODO:
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer, receiver := createTestLexer(bytes.NewBufferString(tc.input))
			err := lexer.Run(context.TODO())
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.want, receiver.Slice)
		})
	}
}

func TestGrammarRandomTestCases(t *testing.T) {
	tests := []*RandomTestCase{}

	testsOpsCount := uint(10)
	testsCount := 10

	for i := 0; i < testsCount; i++ {
		tests = append(tests, NewRandomTestCase(testsOpsCount, true, true))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, NewRandomTestCase(testsOpsCount, false, true))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, NewRandomTestCase(testsOpsCount, true, false))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, NewRandomTestCase(testsOpsCount, false, false))
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lexer, receiver := createTestLexer(tc.Input())
			err := lexer.Run(context.TODO())
			assert.ErrorIs(t, err, io.EOF)
			assert.Len(t, receiver.Slice, tc.tokens)
			if t.Failed() {
				t.Logf("FAILED content: %s", tc.Content())
			}
		})
	}
}

// To run:
// go test -v -run=XXX -bench=BenchmarkGrammar
func BenchmarkGrammar(b *testing.B) {
	type testCase struct {
		name         string
		opsCount     uint
		randomSpaces bool
		randomScopes bool
	}
	tests := []testCase{
		{
			name:         "1e2 ops with random spaces and random scopes",
			opsCount:     1e2,
			randomSpaces: true,
			randomScopes: true,
		},
		{
			name:         "1e3 ops with random spaces and random scopes",
			opsCount:     1e3,
			randomSpaces: true,
			randomScopes: true,
		},
		{
			name:         "1e4 ops with random spaces and random scopes",
			opsCount:     1e4,
			randomSpaces: true,
			randomScopes: true,
		},
		{
			name:         "1e5 ops with random spaces and random scopes",
			opsCount:     1e5,
			randomSpaces: true,
			randomScopes: false,
		},
		{
			name:         "1e6 ops with random spaces and random scopes",
			opsCount:     1e6,
			randomSpaces: true,
			randomScopes: false,
		},
	}
	for _, tc := range tests {
		rtc := NewRandomTestCase(tc.opsCount, tc.randomSpaces, tc.randomScopes)
		lexer := createBenchmarkLexer(rtc.Input())
		b.Run(tc.name, func(b *testing.B) {
			if err := lexer.Run(context.TODO()); !errors.Is(err, io.EOF) {
				b.Fatal("unexpected error")
			}
			b.StopTimer()
			elapsed := b.Elapsed().Seconds()
			tmUnit := "s"
			b.Logf("%s complete in %f%s", tc.name, elapsed, tmUnit)
			b.Logf("\t- %d tokens in %f%s (%f token/%s)", rtc.Tokens(), elapsed, tmUnit, float64(rtc.Tokens())/float64(elapsed), tmUnit)
			b.Logf("\t- %d bytes in %f%s (%f bytes/%s)", rtc.Size(), elapsed, tmUnit, float64(rtc.Size())/float64(elapsed), tmUnit)
		})
	}
}
