package grammar

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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

type randomTestCase struct {
	name    string
	content string
	input   io.Reader
	tokens  int
}

func newRandomTestCase(opsCount uint, randomSpaces bool, randomScopes bool) (ret *randomTestCase) {
	reader, size, tokens := generateRandomInput(opsCount, randomSpaces, randomScopes)
	ret = &randomTestCase{
		name: fmt.Sprintf(
			"%d ops spaces: %t scopes %t size: %d tokens: %d",
			opsCount, randomSpaces, randomScopes, size, tokens,
		),
		input:   reader,
		tokens:  tokens,
		content: reader.String(),
	}
	return ret
}

func (rtc randomTestCase) Name() string {
	return rtc.name
}

func (rtc randomTestCase) Input() io.Reader {
	return rtc.input
}

func (rtc randomTestCase) Tokens() int {
	return rtc.tokens
}

func (rtc randomTestCase) Content() string {
	return rtc.content
}

func TestGrammarRandomTestCases(t *testing.T) {
	tests := []*randomTestCase{}

	testsOpsCount := uint(10)
	testsCount := 10

	for i := 0; i < testsCount; i++ {
		tests = append(tests, newRandomTestCase(testsOpsCount, true, true))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, newRandomTestCase(testsOpsCount, false, true))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, newRandomTestCase(testsOpsCount, true, false))
	}
	for i := 0; i < testsCount; i++ {
		tests = append(tests, newRandomTestCase(testsOpsCount, false, false))
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
		reader, size, tokens := generateRandomInput(tc.opsCount, tc.randomSpaces, tc.randomScopes)
		lexer := createBenchmarkLexer(reader)
		b.Run(tc.name, func(b *testing.B) {
			if err := lexer.Run(context.TODO()); !errors.Is(err, io.EOF) {
				b.Fatal("unexpected error")
			}
			b.StopTimer()
			elapsed := b.Elapsed().Seconds()
			tmUnit := "s"
			b.Logf("%s complete in %f%s", tc.name, elapsed, tmUnit)
			b.Logf("\t- %d tokens in %f%s (%f token/%s)", tokens, elapsed, tmUnit, float64(tokens)/float64(elapsed), tmUnit)
			b.Logf("\t- %d bytes in %f%s (%f bytes/%s)", size, elapsed, tmUnit, float64(size)/float64(elapsed), tmUnit)
		})
	}
}
