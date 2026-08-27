package metric

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	mockStorage := NewMockStorage(t)

	metricService := NewService(mockStorage)

	assert.NotNil(t, metricService)
	assert.Equal(t, metricService.storage, mockStorage)
}

func TestService_GaugeUpdate(t *testing.T) {
	tests := []struct {
		name  string
		value float64
	}{
		{
			name:  "update gauge with positive value",
			value: 42.5,
		},
		{
			name:  "update gauge with zero value",
			value: 0,
		},
		{
			name:  "update gauge with negative value",
			value: -10.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().
				GaugeUpdate("test_gauge", tt.value).
				Return(nil).
				Once()

			metricService := NewService(mockStorage)

			// Act
			err := metricService.GaugeUpdate("test_gauge", tt.value)
			assert.NoError(t, err)

			// Assert
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestService_CounterAdd(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{
			name:  "add counter with positive value",
			value: 42,
		},
		{
			name:  "add counter with zero value",
			value: 0,
		},
		{
			name:  "add counter with negative value",
			value: -100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().CounterAdd("test_counter", tt.value).Return(nil).Once()
			metricService := NewService(mockStorage)
			err := metricService.CounterAdd("test_counter", tt.value)
			mockStorage.AssertExpectations(t)
			assert.NoError(t, err)
		})
	}
}

func TestService_CounterByNameAccumulative(t *testing.T) {

	type given struct {
		name  string
		value int64
		err   error
	}
	type want struct {
		value int64
		err   error
	}
	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "CounterByNameAccumulative with positive value and no error",
			given: given{
				name:  "counter",
				value: 42,
				err:   nil,
			},
			want: want{
				value: 42,
				err:   nil,
			},
		},
		{
			name: "CounterByNameAccumulative with negative value and no error",
			given: given{
				name:  "counter",
				value: -42,
				err:   nil,
			},
			want: want{
				value: -42,
				err:   nil,
			},
		},
		{
			name: "CounterByNameAccumulative with zero value and no error",
			given: given{
				name:  "counter",
				value: 0,
				err:   nil,
			},
			want: want{
				value: 0,
				err:   nil,
			},
		},

		{
			name: "CounterByNameAccumulative with zero value and error",
			given: given{
				name:  "counter",
				value: 0,
				err:   errors.New("error"),
			},
			want: want{
				value: 0,
				err:   errors.New("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().CounterAccumulativeByName(tt.given.name).Return(tt.given.value, tt.given.err).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.CounterByNameAccumulative(tt.given.name)

			mockStorage.AssertExpectations(t)

			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_GaugeByName(t *testing.T) {
	type given struct {
		name  string
		value float64
		err   error
	}
	type want struct {
		value float64
		err   error
	}
	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "GaugeByName with positive value and no error",
			given: given{
				name:  "gauge",
				value: 42.42,
				err:   nil,
			},
			want: want{
				value: 42.42,
				err:   nil,
			},
		},
		{
			name: "GaugeByName with negative value and no error",
			given: given{
				name:  "gauge",
				value: -42.42,
				err:   nil,
			},
			want: want{
				value: -42.42,
				err:   nil,
			},
		},
		{
			name: "GaugeByName with zero value and no error",
			given: given{
				name:  "gauge",
				value: 0,
				err:   nil,
			},
			want: want{
				value: 0,
				err:   nil,
			},
		},
		{
			name: "GaugeByName with zero value and error",
			given: given{
				name:  "gauge",
				value: 0,
				err:   errors.New("error"),
			},
			want: want{
				value: 0,
				err:   errors.New("error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().GaugeByName(tt.given.name).Return(tt.given.value, tt.given.err).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.GaugeByName(tt.given.name)
			mockStorage.AssertExpectations(t)
			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_Gauge(t *testing.T) {
	type want struct {
		value map[string]float64
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "empty map",
			want: want{
				value: map[string]float64{},
			},
		},
		{
			name: "not empty map",
			want: want{
				value: map[string]float64{
					"counter": 42.42,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().Gauge().Return(tt.want.value).Once()
			metricService := NewService(mockStorage)
			value := metricService.Gauge()
			mockStorage.AssertExpectations(t)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_CounterAccumulative(t *testing.T) {
	type want struct {
		value map[string]int64
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "empty map",
			want: want{
				value: map[string]int64{},
			},
		},
		{
			name: "not empty map",
			want: want{
				value: map[string]int64{
					"counter": 42,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().CounterAccumulative().Return(tt.want.value).Once()
			metricService := NewService(mockStorage)
			value := metricService.CounterAccumulative()
			mockStorage.AssertExpectations(t)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_Ping(t *testing.T) {
	t.Run("ping is nil", func(t *testing.T) {
		mockStorage := NewMockStorage(t)
		mockStorage.EXPECT().Ping(nil).Return(nil).Once()
		metricService := NewService(mockStorage)
		err := metricService.Ping(nil)
		assert.Nil(t, err)
	})
}
