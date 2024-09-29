package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVm1(t *testing.T) {
	type testCase struct {
		name    string
		code    []Cell
		want    []Cell
		wantErr error
	}

	tests := []testCase{
		{
			name:    "add 1 cells in stack",
			wantErr: ErrNotEnoughArguments,
			code: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(2)},
			},
		},
		{
			name:    "add 2 cells in stack",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(2)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(4)},
			},
		},
		{
			name:    "add 3 cells in stack",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(2)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(4)},
			},
		},
		{
			name:    "(1+1) (1+1)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(1)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
				{Op: Val, Value: int64(1)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(1)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
				{Op: Val, Value: int64(2)},
			},
		},
		{
			name:    "(1+2) (3+4)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(4)},
				{Op: Val, Value: int64(3)},
				{Op: Add},
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(4)},
				{Op: Val, Value: int64(3)},
				{Op: Add},
				{Op: Val, Value: int64(3)},
			},
		},
		{
			name:    "invalid identifier",
			wantErr: ErrUnknownIdentifier,
			code: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Val, Value: "garbage_ggg"},
				{Op: Call},
			},
			want: []Cell{
				{Op: Val, Value: int64(0)},
			},
		},
		{
			name:    "sin(0)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Ident, Value: "sin"},
				{Op: Call},
			},
			want: []Cell{
				{Op: Val, Value: float64(0)},
			},
		},
		{
			name:    "reset",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Ident, Value: "reset"},
				{Op: Call},
			},
			want: []Cell{},
		},
		{
			name:    "pow(1+1+2+2)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(1)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(2)},
				{Op: Add},
				{Op: Ident, Value: "pow"},
				{Op: Call},
			},
			want: []Cell{
				{Op: Val, Value: float64(16)},
			},
		},
		{
			name:    "set(x,1)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Ident, Value: "x"},
				{Op: Call},
				{Op: Val, Value: int64(1)},
				{Op: Ident, Value: "set"},
				{Op: Call},
			},
			want: []Cell{
				{Op: Val, Value: int64(1)},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := New().PushCode(tc.code)
			err := vm.Run()
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, vm.state.AsSlice())
		})
	}
}
