package csv

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/diakovliev/lexer"
	"github.com/diakovliev/lexer/logger"
	"github.com/diakovliev/lexer/message"
	"github.com/diakovliev/lexer/state"
	"github.com/diakovliev/lexer/xio"
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

func asPtr[T any](v T) *T {
	return &v
}

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

// Receive receives a message from the parser and updates its state accordingly.
func (p *Parser) Receive(msgs []*message.Message[Token]) (err error) {
	for _, msg := range msgs {
		if msg.Type == message.Error {
			err = msg.AsError()
			return
		}
		switch msg.Token {
		case Separator:
			p.currentIndex++
		case Name:
			p.header[p.currentIndex] = msg.AsString()
		case Value:
			p.ensureCurrentIndex(p.currentIndex)
			p.current[p.currentIndex] = msg.AsString()
		case NL:
			p.appendRow()
		default:
			panic("unreachable")
		}
	}
	return
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

func (p *Parser) Grammar(
	nlCb func(context.Context, xio.State) error,
	emitCb func() Token,
) func(b state.Builder[Token]) []state.Update[Token] {
	return func(b state.Builder[Token]) []state.Update[Token] {
		return state.AsSlice[state.Update[Token]](
			// emit separator
			b.Named("Separator").Byte(p.separator).Emit(Separator),
			// generate new lines
			b.Named("NL").Bytes([]byte("\n"), []byte("\r\n")).Emit(NL).Tap(nlCb),
			// generate name or value
			b.Named("NameOrValue").UntilByte(state.Or(
				state.IsByte(p.separator),
				state.IsByte('\n'),
				state.IsByte('\r'),
			)).EmitFn(emitCb),
			// if we are here - emit error
			b.Named("Error").Rest().Error(ErrInvalidInput),
		)
	}
}

func (p *Parser) Parse(input io.Reader) (rows []Row, err error) {
	p.reset()
	var token *Token = asPtr(Value)
	if p.withHeader {
		token = asPtr(Name)
	}
	lex := lexer.New(
		logger.Nop(),
		input,
		message.DefaultFactory[Token](),
		p,
	).With(p.Grammar(
		func(_ context.Context, _ xio.State) (err error) {
			token = asPtr(Value)
			p.resetIndex()
			return
		},
		func() Token {
			return *token
		}))
	err = lex.Run(context.Background())
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	err = nil
	rows = p.complete().Rows
	return
}
