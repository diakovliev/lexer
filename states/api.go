package states

import "github.com/diakovliev/lexer/xio"

type State[T any] interface {
	Update(tx xio.ReadUnreadData) (err error)
}
