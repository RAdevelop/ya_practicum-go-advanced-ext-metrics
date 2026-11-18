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
		signKey               string
		auditURL              string
		auditFile             string
	}

	type want struct {
		address               string
		metricStoreInterval   *time.Duration
		metricRestoreMock     *bool
		metricFileStoragePath string
		signKey               string
		auditURL              string
		auditFile             string
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
				signKey:               "",
				auditFile:             "",
				auditURL:              "",
			},
			want: want{
				address:               "",
				metricStoreInterval:   nil,
				metricRestoreMock:     nil,
				metricFileStoragePath: "",
				signKey:               "",
				auditFile:             "",
				auditURL:              "",
			},
		},
		{
			name: "use ip address",
			given: given{
				address:               "127.0.0.1:9090",
				metricStoreInterval:   new(time.Duration(0)),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(true),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "",
			},
			want: want{
				address:               "127.0.0.1:9090",
				metricStoreInterval:   new(time.Duration(0)),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(true),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "",
			},
		},
		{
			name: "use localhost address",
			given: given{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "",
			},
			want: want{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "",
			},
		},
		{
			name: "auditFile",
			given: given{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "auditFile",
				auditURL:              "",
			},
			want: want{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "auditFile",
				auditURL:              "",
			},
		},
		{
			name: "auditURL",
			given: given{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "auditURL",
			},
			want: want{
				address:               "localhost:9090",
				metricStoreInterval:   new(time.Duration(10) * time.Second),
				metricFileStoragePath: "path",
				metricRestoreMock:     new(false),
				signKey:               "signKey",
				auditFile:             "",
				auditURL:              "auditURL",
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
			mockConfigProvider.EXPECT().SignKey().Return(tt.given.signKey)
			mockConfigProvider.EXPECT().AuditFile().Return(tt.given.auditFile)
			mockConfigProvider.EXPECT().AuditURL().Return(tt.given.auditURL)

			cfg := New(mockConfigProvider)

			assert.Equal(t, tt.want.address, cfg.Address())
			assert.Equal(t, tt.want.metricRestoreMock, cfg.Restore())
			assert.Equal(t, tt.want.metricStoreInterval, cfg.StoreInterval())
			assert.Equal(t, tt.want.metricFileStoragePath, cfg.FileStoragePath())
			assert.Equal(t, tt.want.signKey, cfg.SignKey())
			assert.Equal(t, tt.want.auditFile, cfg.AuditFile())
			assert.Equal(t, tt.want.auditURL, cfg.AuditURL())

			mockConfigProvider.AssertExpectations(t)
		})
	}
}
