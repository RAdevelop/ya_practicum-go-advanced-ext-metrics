package db

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	type want struct {
		envCfg
	}

	tests := []struct {
		name string
		want want
	}{
		{
			name: "no error",
			want: want{
				envCfg: envCfg{
					DSN:             "database_dsn",
					MaxConns:        25,
					MinConns:        5,
					MaxConnLifetime: time.Hour,
					MaxConnIdleTime: time.Minute,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfigProvider := NewMockConfigProvider(t)

			mockConfigProvider.EXPECT().DSN().Return(tt.want.DSN).Once()
			mockConfigProvider.EXPECT().MaxConns().Return(tt.want.MaxConns).Once()
			mockConfigProvider.EXPECT().MinConns().Return(tt.want.MinConns).Once()
			mockConfigProvider.EXPECT().MaxConnLifetime().Return(tt.want.MaxConnLifetime).Once()
			mockConfigProvider.EXPECT().MaxConnIdleTime().Return(tt.want.MaxConnIdleTime).Once()

			dbConf := New(mockConfigProvider)

			assert.Equal(t, tt.want.DSN, dbConf.DSN())
			assert.Equal(t, tt.want.MaxConns, dbConf.MaxConns())
			assert.Equal(t, tt.want.MinConns, dbConf.MinConns())
			assert.Equal(t, tt.want.MaxConnLifetime, dbConf.MaxConnLifetime())
			assert.Equal(t, tt.want.MaxConnIdleTime, dbConf.MaxConnIdleTime())
		})
	}
}
