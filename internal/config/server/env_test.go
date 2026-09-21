package server

import (
	"testing"

	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

func TestEnv_Address(t *testing.T) {

	type want struct {
		address string
		hasErr  bool
	}

	tests := []struct {
		name string
		env  *env.Options
		want want
	}{
		{
			name: "server address not empty",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS": "localhost:8080",
				},
			},
			want: want{
				address: "localhost:8080",
				hasErr:  false,
			},
		},
		{
			name: "server address is empty",
			env: &env.Options{
				Environment: map[string]string{
					"ADDRESS": "",
				},
			},
			want: want{
				address: "",
				hasErr:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envCfg, err := NewEnvWithOptions(tt.env)
			if tt.want.hasErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.address, envCfg.Address())
			}
		})
	}
}
