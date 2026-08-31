package converter

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumericToString(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "int(1)",
			args: args{v: int(1)},
			want: "1",
		},
		{
			name: "int(-1)",
			args: args{v: int(-1)},
			want: "-1",
		},
		{
			name: "int32(1)",
			args: args{v: int32(1)},
			want: "1",
		},
		{
			name: "int32(-1)",
			args: args{v: int32(-1)},
			want: "-1",
		},
		{
			name: "int64(1)",
			args: args{v: int64(1)},
			want: "1",
		},
		{
			name: "int64(-1)",
			args: args{v: int64(-1)},
			want: "-1",
		},
		{
			name: "uint(1)",
			args: args{v: uint(1)},
			want: "1",
		},
		{
			name: "uint32(1)",
			args: args{v: uint32(1)},
			want: "1",
		},
		{
			name: "uint64(1)",
			args: args{v: uint64(1)},
			want: "1",
		},
		{
			name: "math.MaxInt64",
			args: args{v: math.MaxInt64},
			want: "9223372036854775807",
		},
		{
			name: "math.MinInt64",
			args: args{v: math.MinInt64},
			want: "-9223372036854775808",
		},
		{
			name: "float32(123.123)",
			args: args{v: float32(123.123)},
			want: "123.123",
		},
		{
			name: "float32(-123.123)",
			args: args{v: float32(-123.123)},
			want: "-123.123",
		},
		{
			name: "float64(123.123)",
			args: args{v: float64(123.123)},
			want: "123.123",
		},
		{
			name: "float64(-123.123)",
			args: args{v: float64(-123.123)},
			want: "-123.123",
		},
		{
			name: "empty string",
			args: args{v: ""},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumericToString(tt.args.v)
			assert.Equalf(t, got, tt.want, "NumericToString() = %v, want %v", got, tt.want)
		})
	}
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		want    float64
		wantErr bool
	}{
		{
			name:    "int(1)",
			value:   1,
			want:    1,
			wantErr: false,
		},
		{
			name:    "uint(1)",
			value:   uint(1),
			want:    1,
			wantErr: false,
		},
		{
			name:    "uint32(1)",
			value:   uint32(1),
			want:    1,
			wantErr: false,
		},
		{
			name:    "uint64(1)",
			value:   uint64(1),
			want:    1,
			wantErr: false,
		},
		{
			name:    "int32(1)",
			value:   int32(1),
			want:    1,
			wantErr: false,
		},
		{
			name:    "int64(1)",
			value:   int64(1),
			want:    1,
			wantErr: false,
		},
		{
			name:    "float32(123.123)",
			value:   float32(123.123),
			want:    123.123,
			wantErr: false,
		},
		{
			name:    "float64(123.123)",
			value:   123.123,
			want:    123.123,
			wantErr: false,
		},
		{
			name:    "float64(-123.123)",
			value:   -123.123,
			want:    -123.123,
			wantErr: false,
		},
		{
			name:    "string(-123.123)",
			value:   "-123.123",
			want:    -123.123,
			wantErr: false,
		},
		{
			name:    "string()",
			value:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "nil",
			value:   nil,
			want:    0,
			wantErr: true,
		},
		{
			name:    "bool(true)",
			value:   true,
			want:    0,
			wantErr: true,
		},
		{
			name:    "bool(false)",
			value:   false,
			want:    0,
			wantErr: true,
		},
		{
			name:    "struct {}{}",
			value:   struct{}{},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat64(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.InDelta(t, tt.want, got, 1e-5, "ToFloat64(%v) = %v, want %v", tt.value, got, tt.want)
		})
	}
}
