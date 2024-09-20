package state

import (
	"context"

	"github.com/diakovliev/lexer/xio"
)

type State[T any] interface {
	Update(ctx context.Context, tx xio.State) (err error)
}
