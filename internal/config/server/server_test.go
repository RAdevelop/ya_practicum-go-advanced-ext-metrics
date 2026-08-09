package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfigProvider struct {
	address func() string
}

func NewTestConfigProvider(addressMock func() string) *TestConfigProvider {
	return &TestConfigProvider{
		address: addressMock,
	}
}

func (tcp *TestConfigProvider) Address() string {
	return tcp.address()
}

func TestConfig_Address(t *testing.T) {

	tests := []struct {
		name        string
		addressMock func() string
		want        string
	}{
		{
			name: "use empty address",
			addressMock: func() string {
				return ""
			},
			want: "",
		},
		{
			name: "use ip address",
			addressMock: func() string {
				return "127.0.0.1:9090"
			},
			want: "127.0.0.1:9090",
		},
		{
			name: "use localhost address",
			addressMock: func() string {
				return "localhost:9090"
			},
			want: "localhost:9090",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgProvider := NewTestConfigProvider(tt.addressMock)
			cfg := New(cfgProvider)
			assert.Equal(t, tt.want, cfg.Address())
		})
	}
}
