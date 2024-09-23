package state

import (
	"github.com/diakovliev/lexer/common"
	"github.com/diakovliev/lexer/xio"
	"github.com/stretchr/testify/mock"
)

type (
	XioStateMock struct {
		mock.Mock
		newState func() *XioStateMock
	}

	XioSourceMock struct {
		mock.Mock
		newState func() *XioStateMock
	}
)

func newXioStateMock(newState func() *XioStateMock) *XioStateMock {
	return &XioStateMock{
		newState: newState,
	}
}

func (m *XioStateMock) Begin() common.IfaceRef[xio.State] {
	m.Called()
	return common.Ref[xio.State](m.newState())
}
func (m *XioStateMock) Has() bool {
	args := m.Called()
	return args.Bool(0)
}
func (m *XioStateMock) Data() (data []byte, pos int64, err error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Get(1).(int64), args.Error(2)
}
func (m *XioStateMock) NextByte() (b byte, err error) {
	args := m.Called()
	return args.Get(0).(byte), args.Error(1)
}
func (m *XioStateMock) NextRune() (r rune, w int, err error) {
	args := m.Called()
	return args.Get(0).(rune), args.Get(1).(int), args.Error(2)
}
func (m *XioStateMock) Read(p []byte) (n int, err error) {
	args := m.Called()
	return args.Get(0).(int), args.Error(1)
}
func (m *XioStateMock) Unread() (n int, err error) {
	args := m.Called()
	return args.Get(0).(int), args.Error(1)
}
func (m *XioStateMock) Commit() error {
	args := m.Called()
	return args.Error(0)
}
func (m *XioStateMock) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

func newXioSourceMock(newState func() *XioStateMock) *XioSourceMock {
	return &XioSourceMock{
		newState: newState,
	}
}

func (m *XioSourceMock) Begin() common.IfaceRef[xio.State] {
	m.Called()
	return common.Ref[xio.State](m.newState())
}
func (m *XioSourceMock) Has() bool {
	args := m.Called()
	return args.Bool(0)
}
