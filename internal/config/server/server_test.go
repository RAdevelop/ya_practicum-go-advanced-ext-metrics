package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type TestConfProvider struct {
	address               func() string
	metricStoreInterval   func() *time.Duration
	metricFileStoragePath func() string
	metricRestoreMock     func() *bool
}

func NewTestConfigProvider(addressMock func() string, metricStoreInterval func() *time.Duration, metricFileStoragePath func() string, metricRestoreMock func() *bool) *TestConfProvider {
	return &TestConfProvider{
		address:               addressMock,
		metricStoreInterval:   metricStoreInterval,
		metricFileStoragePath: metricFileStoragePath,
		metricRestoreMock:     metricRestoreMock,
	}
}

func (tcp *TestConfProvider) Address() string {
	return tcp.address()
}
func (tcp *TestConfProvider) StoreInterval() *time.Duration {
	return tcp.metricStoreInterval()
}

func (tcp *TestConfProvider) FileStoragePath() string {
	return tcp.metricFileStoragePath()
}

func (tcp *TestConfProvider) Restore() *bool {
	return tcp.metricRestoreMock()
}

func TestConfig(t *testing.T) {

	type want struct {
		address               string
		metricStoreInterval   *time.Duration
		metricRestoreMock     *bool
		metricFileStoragePath string
	}

	tests := []struct {
		name                  string
		addressMock           func() string
		metricStoreInterval   func() *time.Duration
		metricRestoreMock     func() *bool
		metricFileStoragePath func() string
		want                  want
	}{
		{
			name: "use empty",
			addressMock: func() string {
				return ""
			},
			metricStoreInterval: func() *time.Duration {
				return nil
			},
			metricRestoreMock: func() *bool {
				return nil
			},
			metricFileStoragePath: func() string {
				return ""
			},
			want: want{
				address:               "",
				metricStoreInterval:   nil,
				metricRestoreMock:     nil,
				metricFileStoragePath: "",
			},
		},
		{
			name: "use ip address",
			addressMock: func() string {
				return "127.0.0.1:9090"
			},
			metricStoreInterval: func() *time.Duration {
				return new(time.Duration(0))
			},
			metricFileStoragePath: func() string {
				return "path"
			},
			metricRestoreMock: func() *bool {
				return new(true)
			},
			want: want{
				address:               "127.0.0.1:9090",
				metricStoreInterval:   new(time.Duration(0)),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(true),
			},
		},
		{
			name: "use localhost address",
			addressMock: func() string {
				return "localhost:9090"
			},
			metricStoreInterval: func() *time.Duration {
				return new(time.Duration(10) * time.Second)
			},
			metricFileStoragePath: func() string {
				return "path"
			},
			metricRestoreMock: func() *bool {
				return new(false)
			},
			want: want{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgProvider := NewTestConfigProvider(tt.addressMock, tt.metricStoreInterval, tt.metricFileStoragePath, tt.metricRestoreMock)
			cfg := New(cfgProvider)
			assert.Equal(t, tt.want.address, cfg.Address())
			assert.Equal(t, tt.want.metricRestoreMock, cfg.Restore())
			assert.Equal(t, tt.want.metricStoreInterval, cfg.StoreInterval())
			assert.Equal(t, tt.want.metricFileStoragePath, cfg.FileStoragePath())
		})
	}
}
