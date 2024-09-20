package state

import "github.com/diakovliev/lexer/xio"

type State[T any] interface {
	Update(tx xio.State) (err error)
}
