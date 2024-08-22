package lexer

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/diakovliev/vector"
)

var (
	// errEOF is an error that is returned when the end of the input is reached.
	errEOF = errors.New("EOF")
	// ErrLexerError is an error that is returned when an error occurs during the lexing process.
	ErrLexerError = errors.New("lexer error")
	// ErrCallbackError is an error that is returned when an error occurs during the callback function.
	ErrCallbackError = errors.New("callback error")
)

// Lexer is a lexer
type Lexer[T any] struct {
	// Error is the lexer error
	Error error
	// State
	startState State[T]
	// input buffer
	input []byte
	// current line
	line *line
	// current lexeme
	lexeme *lexeme
	// callback
	callback func(Message[T]) error
	// emits history.
	// History keeps historyLen copies of emitted messages.
	// History is a stack, so the most recent message is at the top
	// and the oldest message is at the bottom.
	history *vector.StackImpl[Message[T]]
	// emit history len
	// historyLen == 0 means no history
	// historyLen < 0 means unlimited history
	historyLen int
}

func New[T any](input []byte, startState State[T]) *Lexer[T] {
	return &Lexer[T]{
		startState: startState,
		input:      input,
		line:       &line{},
		lexeme:     &lexeme{},
		history:    vector.NewStack[Message[T]](),
		historyLen: 0,
	}
}

// WithHistory sets the maximum history length for the Lexer.
//
// maxHistoryLen: an integer representing the maximum history length.
// *Lexer[T]: a pointer to the modified Lexer.
func (lex Lexer[T]) WithHistory(historyLen int) *Lexer[T] {
	lex.historyLen = historyLen //nolint:revive
	return &lex
}

// WithCallback sets the callback function for the Lexer.
//
// callback: The function to be called for each item in the Lexer.
// Return: A pointer to the modified Lexer.
func (lex Lexer[T]) WithCallback(callback func(Message[T]) error) *Lexer[T] {
	lex.callback = callback //nolint:revive
	return &lex
}

// IsEOF checks if the lexer has reached the end of the input.
//
// It returns a boolean value indicating if the lexer has reached the end of the input.
func (lex *Lexer[T]) IsEOF() bool {
	if lex.Error != nil {
		return true
	}
	return lex.lexeme.pos() >= len(lex.input)
}

// Next returns the next rune from the lexer input and any error encountered.
//
// It checks if the lexer has reached the end of the input and returns the
// ErrEOF error if true. Otherwise, it decodes the next rune from the lexer's
// input and returns it along with any decoding error. If the decoding of the
// rune resulted in an invalid rune, it returns the ErrInvalidRune error along
// with the rune's position and the input string that caused the error.
// The width of the rune is added to the lexer's lexeme width.
//
// Returns:
// - ret: The next rune from the lexer's input.
// - err: Any error encountered during the decoding of the rune.
func (lex *Lexer[T]) Next() (ret rune, undo func(), err error) {
	if lex.Error != nil {
		err = lex.Error
		return
	}
	undo = lex.lexeme.restore()
	if lex.IsEOF() {
		err = errEOF
		return
	}
	ret, w := utf8.DecodeRune(lex.input[lex.lexeme.pos():])
	if ret == utf8.RuneError {
		err = fmt.Errorf("%w: invalid rune", ErrLexerError)
		lex.Error = err
		return
	}
	lex.lexeme.add(w)
	return
}

// Last returns the last emitted message from the lexer.
//
// It does not take any parameters.
// It returns a pointer to the last emitted value or nil if the history is empty.
func (lex Lexer[T]) Last() (ret *Message[T]) {
	if lex.historyLen == 0 || lex.history.Empty() {
		return
	}
	h := lex.history.Top()
	ret = &h
	return
}

// History returns the history of the Lexer.
//
// It returns a pointer to a StackImpl of Message[T].
func (lex Lexer[T]) History() (ret *vector.Impl[Message[T]]) {
	ret = lex.history.Vector
	return
}

// remember adds the given message to the history of the lexer.
//
// It checks if the maximum history length has been reached and removes the oldest message if necessary.
// The message is then added to the history.
func (lex *Lexer[T]) remember(message Message[T]) {
	if lex.historyLen == 0 {
		return
	}
	// Add to history
	if lex.historyLen > 0 && lex.history.Vector.Len() > lex.historyLen {
		lex.history.Vector.Remove(uint(lex.history.Vector.Len() - 1))
	}
	lex.history.Push(message)
}

// Drop discards the current lexeme and moves the lexer to the next lexeme.
func (lex *Lexer[T]) Drop() (err error) {
	message := Message[T]{
		Type:  Drop,
		Value: lex.Lexeme(),
	}
	lex.remember(message)
	// just reinitialize the lexeme
	lex.lexeme = lex.lexeme.next()
	return
}

func (lex *Lexer[T]) call(message Message[T]) (err error) {
	if lex.callback == nil {
		return
	}
	if err = lex.callback(message); err != nil {
		err = fmt.Errorf("%w: %s", ErrCallbackError, err)
		lex.Error = err
	}
	return
}

// Break reports an error in the lexer and returns a nil StateFn.
// It will set the lexer error if the callback function will return an error.
func (lex *Lexer[T]) Break(errorMessage string) (ret StateFn[T]) {
	message := Message[T]{
		Type:  Error,
		Value: []byte(fmt.Sprintf("%s\nState:\n%s", errorMessage, lex.debugBuffer())),
	}
	if lex.call(message) != nil {
		return
	}
	lex.remember(message)
	return
}

func (lex *Lexer[T]) emit(msgType MessageType, userType T) (err error) {
	var value []byte
	if msgType != NL && msgType != EOF {
		value = lex.Lexeme()
	}
	message := Message[T]{
		Type:     msgType,
		UserType: userType,
		Value:    value,
	}
	if err = lex.call(message); err != nil {
		return
	}
	lex.remember(message)
	if msgType == NL {
		lex.line = lex.line.next(lex.lexeme.pos())
	}
	lex.lexeme = lex.lexeme.next()
	return
}

// Emit emits a token of the specified kind and calls the callback function if it is set.
//
// The kind parameter specifies the kind of token to emit.
// The function returns an error if there was an issue calling the callback.
// It returns nil otherwise.
func (lex *Lexer[T]) Emit(msgType T) (err error) {
	return lex.emit(User, msgType)
}

// NL is a function that is used to emit line character lexeme in the lexer.
// It takes no parameters and returns an error.
func (lex *Lexer[T]) NL() (err error) {
	var userType T
	return lex.emit(NL, userType)
}

// EOF is a function that is used to emit EOF lexeme in the lexer.
// It takes no parameters and returns an error.
func (lex *Lexer[T]) EOF() (err error) {
	var userType T
	return lex.emit(EOF, userType)
}

// Peek returns true if the next character in the Lexer matches the given condition.
//
// The function takes in a `peekFn` parameter which is a function that takes a `rune` and returns a boolean.
// It returns a boolean value indicating whether the condition is true or false.
func (lex *Lexer[T]) Peek(peekFn func(rune) bool) (peeked bool) {
	if lex.Error != nil {
		return
	}
	r, undo, err := lex.Next()
	if errors.Is(err, errEOF) {
		lex.lexeme.reset()
		return
	}
	if err != nil {
		return
	}
	peeked = peekFn(r)
	undo()
	return
}

// AcceptRegexp accepts a regular expression and checks if it matches the current input in the lexer.
//
// It takes a *regexp.Regexp as a parameter and returns a boolean value indicating whether the regular expression was matched or not.
func (lex *Lexer[T]) AcceptRegexp(regexp *regexp.Regexp) (accepted bool) {
	if lex.Error != nil || lex.IsEOF() {
		return
	}
	index := regexp.FindIndex(lex.input[lex.lexeme.pos():])
	if index == nil || len(index) != 2 {
		return
	}
	lex.lexeme.add(index[1] - index[0])
	accepted = true
	return
}

// SkipRegexp skips the next occurrence of the given regular expression in the input.
//
// It takes a *regexp.Regexp as a parameter.
// It returns a boolean value indicating whether a match was found and skipped.
func (lex *Lexer[T]) SkipRegexp(regexp *regexp.Regexp) (skipped bool) {
	if lex.Error != nil || lex.IsEOF() {
		return
	}
	index := regexp.FindIndex(lex.input[lex.lexeme.pos():])
	if index == nil || len(index) != 2 {
		return
	}
	lex.lexeme = lex.lexeme.from(lex.lexeme.pos() + index[1] - index[0])
	skipped = true
	return
}

// PeekRegexp checks if the given regular expression matches the input string starting from the current position.
//
// It takes a regular expression as a parameter and returns a boolean indicating whether the regular expression
// matches or not.
func (lex *Lexer[T]) PeekRegexp(regexp *regexp.Regexp) (peeked bool) {
	if lex.Error != nil || lex.IsEOF() {
		return
	}
	index := regexp.FindIndex(lex.input[lex.lexeme.pos():])
	if index == nil || len(index) != 2 {
		return
	}
	peeked = true
	return
}

// Skip skips any character that satisfies the given condition.
//
// Parameters:
// - skipFn: A function that takes a rune and returns a boolean indicating whether to skip the character or not.
//
// Returns:
// - ok: A boolean value indicating whether a character was skipped or not.
// - err: An error, if any.
func (lex *Lexer[T]) Skip(skipFn func(rune) bool) (skipped bool) {
	if lex.Error != nil {
		return
	}
	r, undo, err := lex.Next()
	if errors.Is(err, errEOF) {
		lex.lexeme.reset()
		return
	}
	if err != nil {
		return
	}
	skipped = skipFn(r)
	if !skipped {
		undo()
		return
	}
	lex.lexeme = lex.lexeme.next()
	return
}

// Accept accepts any rune that satisfies the accept function.
//
// acceptFn is a function that takes a rune as input and returns a boolean value indicating whether the rune is accepted or not.
// The function returns a boolean value indicating whether the accepted rune satisfies the accept function or not.
// It also returns an error if there is any issue while getting the next rune.
func (lex *Lexer[T]) Accept(acceptFn func(rune) bool) (accepted bool) {
	if lex.Error != nil {
		return
	}
	r, undo, err := lex.Next()
	if errors.Is(err, errEOF) {
		lex.lexeme.reset()
		return
	}
	if err != nil {
		return
	}
	accepted = acceptFn(r)
	if !accepted {
		undo()
	}
	return
}

// SkipWhile skips characters while the given condition is true.
//
// skipFn: The function that determines if a character should be skipped.
// skipped: A boolean indicating if any characters were skipped.
func (lex *Lexer[T]) SkipWhile(skipFn func(rune) bool) (skipped bool) {
	if lex.Error != nil {
		return
	}
	for !lex.IsEOF() {
		skip := lex.Skip(skipFn)
		if !skip {
			break
		}
		skipped = skip
	}
	return
}

// SkipAnyFrom skips any of the characters in the given text.
//
// It takes a string parameter `text` which represents the characters to skip.
// It returns a boolean value `ok` indicating if any of the characters were skipped successfully.
// It also returns an error value `err` if there was an error encountered during the skipping process.
func (lex *Lexer[T]) SkipAnyFrom(text string) (skipped bool) {
	if lex.Error != nil {
		return
	}
	skipped = lex.Skip(func(r rune) (skipped bool) {
		for _, t := range text {
			skipped = t == r
			if skipped {
				return
			}
		}
		return
	})
	return
}

// AcceptString accepts a string as input and checks if it matches the next
// characters returned by the Lexer. It returns true if the string matches,
// otherwise it returns false. If an error occurs during the matching process,
// it returns the error.
//
// Parameters:
// - text: The string to be matched.
//
// Return:
// - ok: A boolean indicating if the string matches.
// - err: An error, if any, occurred during the matching process.
func (lex *Lexer[T]) AcceptString(text string) (accepted bool) {
	if lex.Error != nil {
		return
	}
	for _, t := range text {
		r, _, err := lex.Next()
		if errors.Is(err, errEOF) || t != r {
			lex.lexeme.reset()
			return
		}
		if err != nil {
			return
		}
	}
	accepted = true
	return
}

// AcceptAnyFrom checks if the next rune in the lexer matches any of the specified characters.
//
// text: the characters to check against.
// ok: true if the next rune matches any of the characters, false otherwise.
// err: any error that occurred during the operation.
// Returns:
//   - ok: true if the next rune matches any of the characters, false otherwise.
//   - err: any error that occurred during the operation.
func (lex *Lexer[T]) AcceptAnyFrom(text string) (accepted bool) {
	if lex.Error != nil {
		return
	}
	accepted = lex.Accept(func(r rune) (accepted bool) {
		for _, t := range text {
			accepted = t == r
			if accepted {
				return
			}
		}
		return
	})
	return
}

// Lexeme returns the portion of the input that corresponds to the current lexeme.
//
// This function does not take any parameters.
// It returns a byte slice, which represents the portion of the input that corresponds to the current lexeme.
func (lex *Lexer[T]) Lexeme() []byte {
	return lex.input[lex.lexeme.start():lex.lexeme.pos()]
}

// String returns the string representation of the current lexeme.
//
// It does not take any parameters.
// It returns a string.
func (lex *Lexer[T]) String() string {
	return strings.ToValidUTF8(string(lex.Lexeme()), "?")
}

// Do executes the Lexer's state machine.
//
// It iterates through each state function and calls it until a nil state function is returned.
func (lex *Lexer[T]) Do() *Lexer[T] {
	for stateFn := lex.startState.State; stateFn != nil; {
		stateFn = stateFn(lex)
	}
	if lex.Error != nil {
		return lex
	} else if !lex.IsEOF() {
		lex.Error = fmt.Errorf("%w: incomplete lexical analysis", ErrLexerError)
	}
	return lex
}
