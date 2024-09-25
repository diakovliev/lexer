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
	tests := []*RandomTestCase{}

	opsCounts := []uint{1e2, 1e3, 1e4, 1e5, 1e6}
	for _, testsOpsCount := range opsCounts {
		tests = append(tests, NewRandomTestCase(testsOpsCount, true, true))
		tests = append(tests, NewRandomTestCase(testsOpsCount, false, true))
		tests = append(tests, NewRandomTestCase(testsOpsCount, true, false))
		tests = append(tests, NewRandomTestCase(testsOpsCount, false, false))
	}

	for _, tc := range tests {
		lexer := createBenchmarkLexer(tc.Input())
		b.Run(tc.Name(), func(b *testing.B) {
			if err := lexer.Run(context.TODO()); !errors.Is(err, io.EOF) {
				b.Fatal("unexpected error")
			}
			b.StopTimer()
			elapsed := b.Elapsed().Seconds()
			tmUnit := "s"
			b.Logf("%s: %f%s", tc.Name(), elapsed, tmUnit)
			b.Logf("\t- %f token/%s", float64(tc.Tokens())/float64(elapsed), tmUnit)
			b.Logf("\t- %f bytes/%s", float64(tc.Size())/float64(elapsed), tmUnit)
		})
	}
}
