package csv

import (
	"context"
	"encoding/json"
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
)

type Token uint

const (
	NL Token = iota
	Name
	Value
)

type (
	Receiver struct {
		header  map[int]string
		current []string
		Objects []Row
	}
)

func newReceiver() *Receiver {
	return &Receiver{
		header:  make(map[int]string),
		current: make([]string, 0),
		Objects: make([]Row, 0),
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
			hdr, ok := r.header[i]
			if ok {
				column = hdr
			}
		}
		row[column] = value
	}
	r.Objects = append(r.Objects, row)
	r.current = make([]string, 0, len(r.header))
}

func (r *Receiver) Receive(msg *message.Message[Token]) (err error) {
	if msg.Type == message.Error {
		return
	}
	switch msg.Token {
	case Name:
		length := len(r.header)
		content, _ := msg.ValueAsBytes()
		r.header[length] = string(content)
	case Value:
		content, _ := msg.ValueAsBytes()
		r.current = append(r.current, string(content))
	case NL:
		r.appendRow()
	}
	return
}

func (r *Receiver) Complete() *Receiver {
	r.appendRow()
	return r
}

func asPtr[T any](v T) *T {
	return &v
}

func Do(input io.Reader, output io.Writer, separator byte, withHeader bool) (err error) {
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
			b.Named("nl").Bytes([]byte("\n"), []byte("\r\n")).Emit(NL).Tap(func(_ context.Context, _ xio.State) (err error) {
				token = asPtr(Value)
				return
			}),
			b.Named("separator").Byte(separator).Omit(),
			b.Named("value").UntilByte(state.Or(
				state.IsByte(separator),
				state.IsByte('\n'),
				state.IsByte('\r'),
			),
			).EmitFn(func() Token {
				return *token
			}),
			b.Named("error").Rest().Error(errors.New("unexpected input")),
		)
	})
	err = lex.Run(context.Background())
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	// reset io.EOF
	err = json.NewEncoder(output).Encode(receiver.Complete().Objects)
	return
}
