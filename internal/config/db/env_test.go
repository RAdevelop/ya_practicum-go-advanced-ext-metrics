package db

import (
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {

	type want struct {
		envCfg
		hasErr bool
	}

	tests := []struct {
		name string
		env  *env.Options
		want want
	}{
		{
			name: "no error",
			env: &env.Options{
				Environment: map[string]string{
					"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "25",
					"MIN_CONNS":          "5",
					"MAX_CONN_LIFETIME":  "1h",
					"MAX_CONN_IDLE_TIME": "1m",
				},
			},
			want: want{
				hasErr: false,
				envCfg: envCfg{
					DSN:             "database_dsn",
					MaxConns:        25,
					MinConns:        5,
					MaxConnLifetime: time.Hour,
					MaxConnIdleTime: time.Minute,
				},
			},
		},
		{
			name: "empty environment",
			env: &env.Options{
				Environment: map[string]string{},
			},
			want: want{
				hasErr: false,
				envCfg: envCfg{
					DSN:             "",
					MaxConns:        25,
					MinConns:        5,
					MaxConnLifetime: time.Hour,
					MaxConnIdleTime: 4 * time.Minute,
				},
			},
		},
		{
			name: "DSN not set",
			env: &env.Options{
				Environment: map[string]string{
					//"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "25",
					"MIN_CONNS":          "5",
					"MAX_CONN_LIFETIME":  "1h",
					"MAX_CONN_IDLE_TIME": "1m",
				},
			},
			want: want{
				hasErr: false,
				envCfg: envCfg{
					MaxConns:        25,
					MinConns:        5,
					MaxConnLifetime: 1 * time.Hour,
					MaxConnIdleTime: 1 * time.Minute,
				},
			},
		},
		{
			name: "MAX_CONNS is not int32",
			env: &env.Options{
				Environment: map[string]string{
					"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "not int32",
					"MIN_CONNS":          "5",
					"MAX_CONN_LIFETIME":  "1h",
					"MAX_CONN_IDLE_TIME": "1m",
				},
			},
			want: want{
				hasErr: true,
			},
		},
		{
			name: "MIN_CONNS is not int32",
			env: &env.Options{
				Environment: map[string]string{
					"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "25",
					"MIN_CONNS":          "not int32",
					"MAX_CONN_LIFETIME":  "1h",
					"MAX_CONN_IDLE_TIME": "1m",
				},
			},
			want: want{
				hasErr: true,
			},
		},
		{
			name: "MAX_CONN_LIFETIME is not time.Duration",
			env: &env.Options{
				Environment: map[string]string{
					"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "25",
					"MIN_CONNS":          "5",
					"MAX_CONN_LIFETIME":  "not time.Duration",
					"MAX_CONN_IDLE_TIME": "1m",
				},
			},
			want: want{
				hasErr: true,
			},
		},
		{
			name: "MAX_CONN_IDLE_TIME is not time.Duration",
			env: &env.Options{
				Environment: map[string]string{
					"DATABASE_DSN":       "database_dsn",
					"MAX_CONNS":          "25",
					"MIN_CONNS":          "5",
					"MAX_CONN_LIFETIME":  "1h",
					"MAX_CONN_IDLE_TIME": "not time.Duration",
				},
			},
			want: want{
				hasErr: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envConf, err := NewEnvWithOptions(tt.env)

			if tt.want.hasErr {
				assert.Errorf(t, err, "env: %v", tt.env)
				assert.Nil(t, envConf, "env: %v", tt.env)
				return
			}

			assert.NoError(t, err)
			assert.Equalf(t, tt.want.DSN, envConf.DSN(), "want DSN: %v, got: %v", tt.want.DSN, envConf.DSN())
			assert.Equalf(t, tt.want.MaxConns, envConf.MaxConns(), "want MaxConns: %v, got %v", tt.want.MaxConns, envConf.MaxConns())
			assert.Equalf(t, tt.want.MinConns, envConf.MinConns(), "want MinConns: %v, got %v", tt.want.MinConns, envConf.MinConns())
			assert.Equalf(t, tt.want.MaxConnLifetime, envConf.MaxConnLifetime(), "want MaxConnLifetime: %v, got %v", tt.want.MaxConnLifetime, envConf.MaxConnLifetime())
			assert.Equalf(t, tt.want.MaxConnIdleTime, envConf.MaxConnIdleTime(), "want MaxConnIdleTime: %v, got: %v", tt.want.MaxConnIdleTime, envConf.MaxConnIdleTime())
		})
	}
}
