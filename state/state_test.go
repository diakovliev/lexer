package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	type mocks struct {
		xioSource *XioStateMock
		xioState  *XioStateMock
	}

	type testCase struct {
		name      string
		state     func(b Builder[Token]) *Chain[Token]
		setup     func(m *mocks)
		wantError error
	}

	tests := []testCase{
		{
			name: "State Commit",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.State(b, func(b Builder[Token]) []Update[Token] {
					return AsSlice[Update[Token]](
						b.append("ErrCommit", newFixedResultState(errCommit)).Break(),
					)
				})
			},
			setup: func(m *mocks) {
				m.xioSource.On("Begin").Once()
				m.xioState.On("Commit").Return(nil).Once()
			},
			wantError: errCommit,
		},
		{
			name: "State Rollback no more states",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.State(b, func(b Builder[Token]) []Update[Token] {
					return AsSlice[Update[Token]](
						b.append("ErrCommit", newFixedResultState(errRollback)).Break(),
					)
				})
			},
			setup: func(m *mocks) {
				m.xioSource.On("Begin").Once()
				m.xioSource.On("Has").Return(true).Once()
				m.xioState.On("Rollback").Return(nil).Once()
			},
			wantError: ErrIncomplete,
		},
		{
			name: "State Rollback has more data",
			state: func(b Builder[Token]) *Chain[Token] {
				return b.State(b, func(b Builder[Token]) []Update[Token] {
					return AsSlice[Update[Token]](
						b.append("ErrCommit", newFixedResultState(errRollback)).Break(),
					)
				})
			},
			setup: func(m *mocks) {
				m.xioSource.On("Begin").Once()
				m.xioSource.On("Has").Return(false).Once()
				m.xioState.On("Rollback").Return(nil).Once()
			},
			wantError: ErrInvalidInput,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := makeTestDisposeBuilder()
			xioState := newXioStateMock(nil)
			xioSource := newXioStateMock(func() *XioStateMock {
				return xioState
			})
			mocks := &mocks{
				xioSource: xioSource,
				xioState:  xioState,
			}
			if tc.setup != nil {
				tc.setup(mocks)
			}
			err := tc.state(b).Update(context.Background(), xioSource)
			if tc.wantError != nil {
				assert.ErrorIs(t, err, tc.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
