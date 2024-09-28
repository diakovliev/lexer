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

	Receiver struct {
		currentIndex int
		header       map[int]string
		current      []string
		Rows         []Row
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

func newReceiver() *Receiver {
	return &Receiver{
		header:  make(map[int]string),
		current: make([]string, 0),
		Rows:    make([]Row, 0),
	}
}

func (r *Receiver) appendRow() {
	if len(r.current) == 0 {
		return
	}
	row := make(Row)
	for i, value := range r.current {
		column := fmt.Sprintf("col%d", i)
		if len(r.header) > 0 {
			if hdr, ok := r.header[i]; ok {
				column = hdr
			}
		}
		row[column] = value
	}
	r.Rows = append(r.Rows, row)
	r.currentIndex = 0
	r.current = make([]string, 0, len(r.header))
}

func (r *Receiver) ensureCurrentIndex(index int) {
	if index < len(r.current) {
		return
	}
	// grow
	oldCurrent := r.current
	r.current = make([]string, index+1)
	copy(r.current, oldCurrent)
}

func (r *Receiver) Receive(msg *message.Message[Token]) (err error) {
	if msg.Type == message.Error {
		err = msg.AsError()
		return
	}
	switch msg.Token {
	case Separator:
		r.currentIndex++
	case Name:
		r.header[r.currentIndex] = msg.AsString()
	case Value:
		r.ensureCurrentIndex(r.currentIndex)
		r.current[r.currentIndex] = msg.AsString()
	case NL:
		r.appendRow()
	default:
		panic("unreachable")
	}
	return
}

func (r *Receiver) Complete() *Receiver {
	// we have to append last row
	r.appendRow()
	r.currentIndex = 0
	return r
}

func (r *Receiver) ResetIndex() {
	r.currentIndex = 0
}

func asPtr[T any](v T) *T {
	return &v
}

func Parse(input io.Reader, separator byte, withHeader bool) (rows []Row, err error) {
	var token *Token = asPtr(Name)
	if !withHeader {
		token = asPtr(Value)
	}
	receiver := newReceiver()
	lex := lexer.New(
		logger.Nop(),
		input,
		message.DefaultFactory[Token](),
		receiver,
	).With(func(b state.Builder[Token]) []state.Update[Token] {
		return state.AsSlice[state.Update[Token]](
			// emit separator
			b.Named("Separator").Byte(separator).Emit(Separator),
			// generate new lines
			b.Named("NL").Bytes([]byte("\n"), []byte("\r\n")).Emit(NL).
				Tap(func(_ context.Context, _ xio.State) (err error) {
					token = asPtr(Value)
					receiver.ResetIndex()
					return
				}),
			// generate name or value
			b.Named("NameOrValue").UntilByte(state.Or(
				state.IsByte(separator),
				state.IsByte('\n'),
				state.IsByte('\r'),
			)).EmitFn(func() Token {
				return *token
			}),
			// if we are here - emit error
			b.Named("Error").Rest().Error(ErrInvalidInput),
		)
	})
	err = lex.Run(context.Background())
	if err != nil && !errors.Is(err, io.EOF) {
		return
	}
	err = nil
	rows = receiver.Complete().Rows
	return
}
