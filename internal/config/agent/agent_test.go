package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {

	type given struct {
		addressMock        string
		reportIntervalMock uint
		pollIntervalMock   uint
		key                string
	}

	type want struct {
		address        string
		reportInterval uint
		pollInterval   uint
		key            string
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "use empty settings",
			given: given{
				addressMock:        "",
				reportIntervalMock: 0,
				pollIntervalMock:   0,
				key:                "",
			},
			want: want{
				address:        "",
				reportInterval: 0,
				pollInterval:   0,
				key:            "",
			},
		},
		{
			name: "use ip address",
			given: given{
				addressMock:        "127.0.0.1:9090",
				reportIntervalMock: 5,
				pollIntervalMock:   2,
				key:                "key",
			},
			want: want{
				address:        "127.0.0.1:9090",
				reportInterval: 5,
				pollInterval:   2,
				key:            "key",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfgProvider := NewMockConfigProvider(t)
			cfgProvider.EXPECT().Address().Return(tt.given.addressMock)
			cfgProvider.EXPECT().ReportInterval().Return(tt.given.reportIntervalMock)
			cfgProvider.EXPECT().PollInterval().Return(tt.given.pollIntervalMock)
			cfgProvider.EXPECT().SignKey().Return(tt.given.key)

			agentConfig := New(cfgProvider)
			assert.Equal(t, tt.want.address, agentConfig.Address())
			assert.Equal(t, tt.want.reportInterval, agentConfig.ReportInterval())
			assert.Equal(t, tt.want.pollInterval, agentConfig.PollInterval())
			assert.Equal(t, tt.want.key, agentConfig.SignKey())
		})
	}
}
