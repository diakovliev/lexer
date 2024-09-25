package grammar

import (
	"bytes"
	"io"
	"math/rand"
	"os"
	"strconv"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
)

func createTestLexer(reader io.Reader) (lex *lexer.Lexer[Token], receiver *message.SliceReceiver[Token]) {
	receiver = message.Slice[Token]()
	lex = lexer.New(
		logger.New(
			logger.WithLevel(logger.Trace),
			logger.WithWriter(os.Stderr),
		),
		reader,
		message.DefaultFactory[Token](),
		receiver,
	).With(BuildState(true))
	return
}

func createBenchmarkLexer(reader io.Reader) (lex *lexer.Lexer[Token]) {
	lex = lexer.New(
		logger.New(
			logger.WithLevel(logger.Trace),
			logger.WithWriter(os.Stderr),
		),
		reader,
		message.DefaultFactory[Token](),
		message.Dispose[Token](),
	).With(BuildState(true))
	return
}

func generateRandomSpaces(dest *bytes.Buffer, n int, enabled bool) {
	if !enabled {
		return
	}
	count := rand.Intn(n)
	for i := 0; i < count; i++ {
		dest.WriteRune(' ')
	}
}

func randomScopeOpen(dest *bytes.Buffer, enabled bool) (opened bool) {
	if !enabled {
		return
	}
	if rand.Intn(2) == 0 {
		return
	}
	dest.WriteRune('(')
	opened = true
	return
}

func scopeClose(dest *bytes.Buffer) {
	dest.WriteRune(')')
}

func generateRandomInput(opsCount uint, enableRandomSpaces bool, enableRandomScopes bool) (reader *bytes.Buffer, size int, tokens int) {
	ops := []string{"+", "-", "*", "/"}
	buffer := bytes.NewBuffer(nil)
	// preallocate some space
	buffer.Grow(int(opsCount * 10))
	scopes := uint(0)
	maxUint := 10000
	// Up to 10 random spaces before first number
	generateRandomSpaces(buffer, 10, enableRandomSpaces)
	for i := uint(0); i < opsCount-1; i++ {
		// randomly open or close scope
		if randomScopeOpen(buffer, enableRandomScopes) {
			tokens++
			scopes++
		} else if scopes > 0 && rand.Intn(2) == 0 {
			// Up to 10 random spaces before before operator
			generateRandomSpaces(buffer, 10, enableRandomSpaces)
			// Operator
			buffer.WriteString(ops[rand.Intn(len(ops))])
			tokens++
			i++
			// Up to 10 random spaces before after operator
			generateRandomSpaces(buffer, 10, enableRandomSpaces)
			// Close scope
			scopeClose(buffer)
			tokens++
			scopes--
		}
		// Up to 10 random spaces before before number
		generateRandomSpaces(buffer, 10, enableRandomSpaces)
		buffer.WriteString(strconv.Itoa(rand.Intn(maxUint)))
		tokens++
		// Up to 10 random spaces before after number
		generateRandomSpaces(buffer, 10, enableRandomSpaces)
		// Operator
		buffer.WriteString(ops[rand.Intn(len(ops))])
		tokens++
		// Up to 10 random spaces before after operator
		generateRandomSpaces(buffer, 10, enableRandomSpaces)
	}
	// Up to 10 random spaces before last number
	generateRandomSpaces(buffer, 10, enableRandomSpaces)
	buffer.WriteString(strconv.Itoa(rand.Intn(maxUint)))
	tokens++
	for scopes > 0 {
		// Up to 10 random spaces before )
		generateRandomSpaces(buffer, 10, enableRandomSpaces)
		scopeClose(buffer)
		tokens++
		scopes--
		// Up to 10 random spaces after )
		generateRandomSpaces(buffer, 10, enableRandomSpaces)
	}
	reader = buffer
	size = buffer.Len()
	return
}
