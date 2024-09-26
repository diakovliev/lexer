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

	// RunePredicate is a function that takes rune and returns true if it should be accepted.
	RunePredicate func(rune) bool

	// BytePredicate is a function that takes byte and returns true if it should be accepted.
	BytePredicate func(byte) bool

	fnMode uint
)

const (
	fnAccept fnMode = iota
	fnLook
)
