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
			name:    "add 1 cell in stack",
			wantErr: ErrNotEnoughArguments,
			code: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Add},
			},
			want: []Cell{
				{Op: Val, Value: int64(2)},
				{Op: Add},
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
				{Op: Val, Value: int64(2)},
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
				{Op: Val, Value: int64(7)},
				{Op: Val, Value: int64(3)},
			},
		},
		{
			name:    "invalid identifier",
			wantErr: ErrUnknownIdentifier,
			code: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Call, Value: "garbage_ggg"},
			},
			want: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Call, Value: "garbage_ggg"},
			},
		},
		{
			name:    "sin(0)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(0)},
				{Op: Call, Value: "sin"},
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
				{Op: Call, Value: "reset"},
			},
			want: []Cell{},
		},
		{
			name:    "pow(1+1,2+2)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Val, Value: int64(1)},
				{Op: Val, Value: int64(1)},
				{Op: Add},
				{Op: Val, Value: int64(2)},
				{Op: Val, Value: int64(2)},
				{Op: Add},
				{Op: Call, Value: "pow"},
			},
			want: []Cell{
				{Op: Val, Value: float64(16)},
			},
		},
		{
			name:    "set(x,1)",
			wantErr: ErrHalt,
			code: []Cell{
				{Op: Call, Value: "x"},
				{Op: Val, Value: int64(1)},
				{Op: Call, Value: "set"},
			},
			want: []Cell{
				{Op: Val, Value: int64(1)},
			},
		},
		{
			name:    "ls-la",
			wantErr: ErrUnknownIdentifier,
			code: []Cell{
				{Op: Call, Value: "ls"},
				{Op: Call, Value: "la"},
				{Op: Sub},
			},
			want: []Cell{
				{Op: Call, Value: "ls"},
				{Op: Call, Value: "la"},
				{Op: Sub},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			vm := New().PushCode(tc.code)
			err := vm.Run()
			assert.ErrorIs(t, err, tc.wantErr)
			assert.Equal(t, tc.want, vm.code.AsSlice())
		})
	}
}
