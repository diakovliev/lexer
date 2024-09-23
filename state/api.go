package state

import (
	"context"

	"github.com/diakovliev/lexer/xio"
)

type Update[T any] interface {
	Update(ctx context.Context, tx xio.State) (err error)
}
