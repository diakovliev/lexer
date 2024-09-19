package states

import (
	"github.com/diakovliev/lexer/common"
)

type State[T any] interface {
	Update(tx common.ReadUnreadData) (err error)
}
