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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NumericToString(tt.args.v)
			assert.Equalf(t, got, tt.want, "NumericToString() = %v, want %v", got, tt.want)
		})
	}
}
