package lexer

import (
	"fmt"
	"strings"
)

func (lex *Lexer[T]) lineEnd() int {
	var counter int
	for _, ch := range lex.input[lex.line.start():] {
		if ch == '\n' {
			return lex.line.start() + counter
		}
		counter++
	}
	return len(lex.input)
}

func (lex *Lexer[T]) debugPos() (debugPos int, debugLine string) {
	eol := lex.lineEnd()
	for pos, ch := range lex.input[lex.line.start():eol] {
		switch ch {
		case '\t':
			if pos+lex.line.start() < lex.lexeme.pos() {
				debugPos += 2
			}
			debugLine += `\t`
		case '\n':
			if pos+lex.line.start() < lex.lexeme.pos() {
				debugPos += 2
			}
			debugLine += `\n`
		case '\r':
			if pos+lex.line.start() < lex.lexeme.pos() {
				debugPos += 2
			}
			debugLine += `\r`
		default:
			if pos+lex.line.start() < lex.lexeme.pos() {
				debugPos++
			}
			debugLine += string(ch)
		}
	}
	debugPos--
	return
}

func (lex *Lexer[T]) debugBuffer() string {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(" start: %d\n", lex.lexeme.start()))
	builder.WriteString(fmt.Sprintf(" width: %d\n", lex.lexeme.width()))
	builder.WriteString(fmt.Sprintf(" pos: %d\n", lex.lexeme.pos()))
	builder.WriteString(fmt.Sprintf(" lexeme: %#v\n", lex.Lexeme()))
	builder.WriteString(fmt.Sprintf(" string: '%s'\n", lex.String()))
	debugPos, debugLine := lex.debugPos()
	builder.WriteString(fmt.Sprintf(" line (%d): %s\n", lex.line.number(), debugLine))
	if debugPos >= 0 {
		builder.WriteString(fmt.Sprintf("           %s%s\n", strings.Repeat("-", debugPos), "^"))
	} else {
		builder.WriteString(fmt.Sprintf("          %s\n", "^"))
	}
	return builder.String()
}
