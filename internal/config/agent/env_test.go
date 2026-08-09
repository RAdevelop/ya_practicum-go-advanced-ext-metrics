package agent

import (
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {

	type want struct {
		addr           string
		intervalReport uint
		intervalPoll   uint
		hasErr         bool
	}

	tests := []struct {
		name string
		env  *env.Options
		want want
	}{
		{
			name: "env with no error",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS":         "localhost:8080",
					"REPORT_INTERVAL": "10",
					"POLL_INTERVAL":   "2",
				},
			},
			want: want{
				addr:           "localhost:8080",
				intervalReport: 10,
				intervalPoll:   2,
				hasErr:         false,
			},
		},
		{
			name: "env with error for REPORT_INTERVAL",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS":         "localhost:8080",
					"REPORT_INTERVAL": "-10",
					"POLL_INTERVAL":   "2",
				},
			},
			want: want{
				hasErr: true,
			},
		},
		{
			name: "env with error for POLL_INTERVAL",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS":         "localhost:8080",
					"REPORT_INTERVAL": "10",
					"POLL_INTERVAL":   "-2",
				},
			},
			want: want{
				hasErr: true,
			},
		},
		{
			name: "env with empty values",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS":         "",
					"REPORT_INTERVAL": "",
					"POLL_INTERVAL":   "",
				},
			},
			want: want{
				addr:           "",
				intervalReport: 0,
				intervalPoll:   0,
				hasErr:         false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgEnv, err := NewEnvWithOptions(tt.env)

			if tt.want.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.addr, cfgEnv.Address())
				assert.Equal(t, tt.want.intervalPoll, cfgEnv.PollInterval())
				assert.Equal(t, tt.want.intervalReport, cfgEnv.ReportInterval())
			}
		})
	}
}
