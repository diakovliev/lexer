package state

import (
	"context"

	"github.com/diakovliev/lexer/xio"
)

type (
	// Update is an interface for updating state.
	Update[T any] interface {
		// Update updates state.
		Update(ctx context.Context, tx xio.State) (err error)
	}

	// fnMode is mode of Fn states
	fnMode uint
)

const (
	// fnAccept in this mode Fn state will advance
	// io state on successful match
	fnAccept fnMode = iota
	// ftLook in this mode FnState will not
	// advance io state on successful match, but
	// will return ErrCommit to the caller. It
	// will allow to look ahead and decide what
	// to do next.
	fnLook
)
