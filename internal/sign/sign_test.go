package sign

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSHA256(t *testing.T) {
	type given struct {
		str       []byte
		secretKey string
	}
	tests := []struct {
		name  string
		given given
		want  string
	}{
		{
			name: "empty key",
			given: given{
				str:       []byte("str"),
				secretKey: "",
			},
			want: "d966114df9e8c208b8e2577e29ca716d3d13f511f5c12e9a9137af917d4d4b1d",
		},
		{
			name: "not empty key",
			given: given{
				str:       []byte("str"),
				secretKey: "secretKey",
			},
			want: "97f5f0af93f954f8ec5e8f48b52f8d539ce90df241529394a1ea6528ce3bb41c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SHA256(tt.given.str, tt.given.secretKey))
		})
	}
}

func TestSHA256Verify(t *testing.T) {
	type given struct {
		str       []byte
		secretKey string
		sign      string
	}
	tests := []struct {
		name  string
		given given
		want  bool
	}{
		{
			name: "empty key",
			given: given{
				str:       []byte("str"),
				secretKey: "",
				sign:      "d966114df9e8c208b8e2577e29ca716d3d13f511f5c12e9a9137af917d4d4b1d",
			},
			want: true,
		},
		{
			name: "not empty key",
			given: given{
				str:       []byte("str"),
				secretKey: "secretKey",
				sign:      "97f5f0af93f954f8ec5e8f48b52f8d539ce90df241529394a1ea6528ce3bb41c",
			},
			want: true,
		},
		{
			name: "not empty key",
			given: given{
				str:       []byte("str"),
				secretKey: "secretKey",
				sign:      "d966114df9e8c208b8e2577e29ca716d3d13f511f5c12e9a9137af917d4d4b1d",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SHA256Verify(tt.given.str, tt.given.secretKey, tt.given.sign))
		})
	}
}
