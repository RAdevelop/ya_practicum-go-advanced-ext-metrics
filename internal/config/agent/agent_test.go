package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfigProvider struct {
	addressMock        func() string
	reportIntervalMock func() uint
	pollIntervalMock   func() uint
}

func NewTestConfigProvider(addressMock func() string, reportIntervalMock func() uint, pollIntervalMock func() uint) *TestConfigProvider {
	return &TestConfigProvider{
		addressMock:        addressMock,
		reportIntervalMock: reportIntervalMock,
		pollIntervalMock:   pollIntervalMock,
	}
}

func (tcp *TestConfigProvider) ReportInterval() uint {
	return tcp.reportIntervalMock()
}

func (tcp *TestConfigProvider) PollInterval() uint {
	return tcp.pollIntervalMock()
}

func (tcp *TestConfigProvider) Address() string {
	return tcp.addressMock()
}

func TestConfig(t *testing.T) {

	type want struct {
		address        string
		reportInterval uint
		pollInterval   uint
	}

	tests := []struct {
		name               string
		addressMock        func() string
		reportIntervalMock func() uint
		pollIntervalMock   func() uint
		want               want
	}{
		{
			name: "use empty settings",
			addressMock: func() string {
				return ""
			},
			reportIntervalMock: func() uint {
				return 0
			},
			pollIntervalMock: func() uint {
				return 0
			},
			want: want{
				address:        "",
				reportInterval: 0,
				pollInterval:   0,
			},
		},
		{
			name: "use ip address",
			addressMock: func() string {
				return "127.0.0.1:9090"
			},
			reportIntervalMock: func() uint {
				return 5
			},
			pollIntervalMock: func() uint {
				return 2
			},
			want: want{
				address:        "127.0.0.1:9090",
				reportInterval: 5,
				pollInterval:   2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			cfgProvider := NewTestConfigProvider(tt.addressMock, tt.reportIntervalMock, tt.pollIntervalMock)
			agentConfig := New(cfgProvider)
			assert.Equal(t, tt.want.address, agentConfig.Address())
			assert.Equal(t, tt.want.reportInterval, agentConfig.ReportInterval())
			assert.Equal(t, tt.want.pollInterval, agentConfig.PollInterval())
		})
	}
}
