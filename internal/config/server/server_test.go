package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {

	type given struct {
		address               string
		metricStoreInterval   *time.Duration
		metricRestoreMock     *bool
		metricFileStoragePath string
	}

	type want struct {
		address               string
		metricStoreInterval   *time.Duration
		metricRestoreMock     *bool
		metricFileStoragePath string
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "use empty",
			given: given{
				address:               "",
				metricStoreInterval:   nil,
				metricRestoreMock:     nil,
				metricFileStoragePath: "",
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
			given: given{
				address:               "127.0.0.1:9090",
				metricStoreInterval:   new(time.Duration(0)),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(true),
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
			given: given{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
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

			mockConfigProvider := NewMockConfigProvider(t)
			mockConfigProvider.EXPECT().Address().Return(tt.given.address)
			mockConfigProvider.EXPECT().Restore().Return(tt.given.metricRestoreMock)
			mockConfigProvider.EXPECT().StoreInterval().Return(tt.given.metricStoreInterval)
			mockConfigProvider.EXPECT().FileStoragePath().Return(tt.given.metricFileStoragePath)

			cfg := New(mockConfigProvider)

			assert.Equal(t, tt.want.address, cfg.Address())
			assert.Equal(t, tt.want.metricRestoreMock, cfg.Restore())
			assert.Equal(t, tt.want.metricStoreInterval, cfg.StoreInterval())
			assert.Equal(t, tt.want.metricFileStoragePath, cfg.FileStoragePath())

			mockConfigProvider.AssertExpectations(t)
		})
	}
}
