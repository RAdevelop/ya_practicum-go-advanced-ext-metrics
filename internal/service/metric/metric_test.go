package metric

import (
	"context"
	"errors"
	"testing"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
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
				GaugeUpdate(context.TODO(), "test_gauge", tt.value).
				Return(nil).
				Once()

			metricService := NewService(mockStorage)

			// Act
			err := metricService.GaugeUpdate(context.TODO(), "test_gauge", tt.value)
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
			mockStorage.EXPECT().CounterAdd(context.TODO(), "test_counter", tt.value).Return(nil).Once()
			metricService := NewService(mockStorage)
			err := metricService.CounterAdd(context.TODO(), "test_counter", tt.value)
			mockStorage.AssertExpectations(t)
			assert.NoError(t, err)
		})
	}
}

func TestService_CounterByNameAccumulative(t *testing.T) {

	type given struct {
		name  string
		value *models.Metrics
		err   error
	}
	type want struct {
		value *models.Metrics
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
				name: "counter",
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(42)),
				},
				err: nil,
			},
			want: want{
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(42)),
				},
				err: nil,
			},
		},
		{
			name: "CounterByNameAccumulative with negative value and no error",
			given: given{
				name: "counter",
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(-42)),
				},
				err: nil,
			},
			want: want{
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(-42)),
				},
				err: nil,
			},
		},
		{
			name: "CounterByNameAccumulative with zero value and no error",
			given: given{
				name: "counter",
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(0)),
				},
				err: nil,
			},
			want: want{
				value: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(0)),
				},
				err: nil,
			},
		},
		{
			name: "CounterByNameAccumulative with zero value and error",
			given: given{
				name:  "counter",
				value: nil,
				err:   errors.New("error"),
			},
			want: want{
				value: nil,
				err:   errors.New("error"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().CounterAccumulativeByName(context.TODO(), tt.given.name).Return(tt.given.value, tt.given.err).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.CounterByNameAccumulative(context.TODO(), tt.given.name)

			mockStorage.AssertExpectations(t)

			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_GaugeByName(t *testing.T) {
	type given struct {
		name  string
		value *models.Metrics
		err   error
	}
	type want struct {
		value *models.Metrics
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
				name: "gauge",
				value: &models.Metrics{
					ID:    "gauge",
					MType: models.Gauge,
					Value: new(float64(42.42)),
				},
				err: nil,
			},
			want: want{
				value: &models.Metrics{
					ID:    "gauge",
					MType: models.Gauge,
					Value: new(42.42),
				},
				err: nil,
			},
		},
		{
			name: "GaugeByName with negative value and no error",
			given: given{
				name: "gauge",
				value: &models.Metrics{
					ID:    "gauge",
					MType: models.Gauge,
					Value: new(-42.42),
				},
				err: nil,
			},
			want: want{
				value: &models.Metrics{
					ID:    "gauge",
					MType: models.Gauge,
					Value: new(-42.42),
				},
				err: nil,
			},
		},
		{
			name: "GaugeByName with zero value and no error",
			given: given{
				name:  "gauge",
				value: nil,
				err:   nil,
			},
			want: want{
				value: nil,
				err:   nil,
			},
		},
		{
			name: "GaugeByName with zero value and error",
			given: given{
				name:  "gauge",
				value: nil,
				err:   errors.New("error"),
			},
			want: want{
				value: nil,
				err:   errors.New("error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().GaugeByName(context.TODO(), tt.given.name).Return(tt.given.value, tt.given.err).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.GaugeByName(context.TODO(), tt.given.name)
			mockStorage.AssertExpectations(t)
			assert.Equal(t, tt.want.err, err)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_Gauge(t *testing.T) {
	type want struct {
		value []models.Metrics
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "empty map",
			want: want{
				value: []models.Metrics{},
			},
		},
		{
			name: "not empty map",
			want: want{
				value: []models.Metrics{
					{
						MType: models.Gauge,
						ID:    "counter",
						Value: new(42.42),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().Gauge(context.TODO()).Return(tt.want.value, nil).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.Gauge(context.TODO())
			mockStorage.AssertExpectations(t)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.value, value)
		})
	}
}

func TestService_CounterAccumulative(t *testing.T) {
	type want struct {
		value []models.Metrics
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "empty map",
			want: want{
				value: []models.Metrics{},
			},
		},
		{
			name: "not empty map",
			want: want{
				value: []models.Metrics{
					{
						ID:    "counter",
						Delta: new(int64(42)),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := NewMockStorage(t)
			mockStorage.EXPECT().CounterAccumulative(context.TODO()).Return(tt.want.value, nil).Once()
			metricService := NewService(mockStorage)
			value, err := metricService.CounterAccumulative(context.TODO())
			mockStorage.AssertExpectations(t)
			assert.NoError(t, err)
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
