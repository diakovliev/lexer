package csv

import (
	"errors"
	"fmt"
	"io"

	"github.com/diakovliev/lexer/iterator"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
)

type (
	Row map[string]string

	Token uint

	Parser struct {
		currentIndex int
		header       map[int]string
		current      []string
		Rows         []Row
		withHeader   bool
		separator    byte
	}
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

const (
	NL Token = iota
	Name
	Value
	Separator
)

func New(opts ...Option) (ret *Parser) {
	ret = &Parser{
		header:  make(map[int]string),
		current: make([]string, 0),
		Rows:    make([]Row, 0),
	}
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

func (p *Parser) appendRow() {
	if len(p.current) == 0 {
		return
	}
	row := make(Row)
	for i, value := range p.current {
		column := fmt.Sprintf("col%d", i)
		if len(p.header) > 0 {
			if hdr, ok := p.header[i]; ok {
				column = hdr
			}
		}
		row[column] = value
	}
	p.Rows = append(p.Rows, row)
	p.currentIndex = 0
	p.current = make([]string, 0, len(p.header))
}

func (p *Parser) ensureCurrentIndex(index int) {
	if index < len(p.current) {
		return
	}
	// grow
	oldCurrent := p.current
	p.current = make([]string, index+1)
	copy(p.current, oldCurrent)
}

func (p *Parser) complete() *Parser {
	// we have to append last row
	p.appendRow()
	p.currentIndex = 0
	return p
}

func (p *Parser) resetIndex() {
	p.currentIndex = 0
}

func (p *Parser) reset() {
	p.header = make(map[int]string)
	p.current = make([]string, 0)
	p.Rows = make([]Row, 0)
}

func (p *Parser) Grammar(b state.Builder[Token]) []state.Update[Token] {
	return state.AsSlice[state.Update[Token]](
		// emit separator
		b.Named("Separator").Byte(p.separator).Emit(Separator),
		// generate new lines
		b.Named("NL").Bytes([]byte("\n"), []byte("\r\n")).Emit(NL),
		// generate name or value
		b.Named("NameOrValue").UntilByte(state.Or(
			state.IsByte(p.separator),
			state.IsByte('\n'),
			state.IsByte('\r'),
		)).Emit(Value),
		// if we are here - emit error
		b.Named("Error").Rest().Error(ErrInvalidInput),
	)
}

func (p *Parser) Parse(input io.Reader) (rows []Row, err error) {
	p.reset()
	header := p.withHeader
	iter := iterator.New[Token](input).With(p.Grammar)
	for msg := range iter.Range {
		if msg.Type == message.Error {
			err = msg.AsError()
			return
		}
		switch msg.Token {
		case Separator:
			p.currentIndex++
		case Value:
			if header {
				p.header[p.currentIndex] = msg.AsString()
			} else {
				p.ensureCurrentIndex(p.currentIndex)
				p.current[p.currentIndex] = msg.AsString()
			}
		case NL:
			p.appendRow()
			p.resetIndex()
			header = false
		default:
			panic("unreachable")
		}
	}
	err = iter.Error
	rows = p.complete().Rows
	return
}
