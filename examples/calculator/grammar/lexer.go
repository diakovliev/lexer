package grammar

import (
	"io"
	"os"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
)

const (
	// The maxScopesDepth is maximum allowed depth of the scopes.
	// We are not lisp so we don't need infinite depth here.
	// If you are crazy, you can replace it by math.MaxUint.
	maxScopesDepth = 30

	// historyDepth is a lexer history depth.
	// We need to look back to one token
	// to be able to parse the expressions with
	// negative/positive numbers.
	historyDepth = 1
)

// New creates a new lexer.
func New(reader io.Reader, receiver message.Receiver[Token]) *lexer.Lexer[Token] {
	return lexer.New(
		logger.New(
			logger.WithLevel(logger.Trace),
			logger.WithWriter(os.Stderr),
		),
		reader,
		message.DefaultFactory[Token](),
		receiver,
		lexer.WithHistoryDepth[Token](historyDepth),
	).With(newState(true, maxScopesDepth))
}
