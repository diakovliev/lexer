package lexer

import (
	"errors"
	"fmt"
)

var ErrInvalidRune = errors.New("invalid rune")

type ErrInvalidRuneData struct {
	err  error
	from int
	to   int
}

func newErrInvalidRune(from, to int) error {
	return &ErrInvalidRuneData{
		err:  ErrInvalidRune,
		from: from,
		to:   to,
	}
}

func (e ErrInvalidRuneData) Error() string {
	return fmt.Errorf("%w: from %d to %d", e.err, e.from, e.to).Error()
}

func (e ErrInvalidRuneData) Unwrap() error {
	return e.err
}

func (e ErrInvalidRuneData) From() int {
	return e.from
}

func (e ErrInvalidRuneData) To() int {
	return e.to
}

func InvalidRuneRange(err error) (from, to int) {
	switch e := err.(type) {
	case ErrInvalidRuneData:
		return e.From(), e.To()
	default:
		return 0, 0
	}
}
