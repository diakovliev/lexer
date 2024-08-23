package lexer

type lexeme2 struct {
	pos  int64
	data []byte
	str  string
}

func (l lexeme2) Pos() int64 {
	return l.pos
}

func (l lexeme2) Width() int {
	return len(l.data)
}

func (l *lexeme2) add(d []byte, r rune) {
	l.data = append(l.data, d...)
	l.str += string(r)
}
