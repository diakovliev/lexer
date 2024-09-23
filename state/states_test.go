package state

import (
	"context"
	"errors"

	"github.com/diakovliev/lexer/xio"
)

type (
	fixResultState struct {
		Result error
	}
	notAState struct{}
)

func (t fixResultState) Update(_ context.Context, _ xio.State) error {
	return t.Result
}

func newNotAState() any {
	return &notAState{}
}

func newFixedResultState(result error) func() any {
	return func() any {
		return &fixResultState{
			Result: result,
		}
	}
}

func newFakeState() any {
	return &fixResultState{
		Result: errors.New("test error"),
	}
}
