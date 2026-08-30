package repository

import (
	"context"
	"testing"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func cleanupStorage(t *testing.T, storage *MemStorage) {
	t.Cleanup(func() {
		storage = nil
	})
}

func TestMemStorage_NewMemStorage(t *testing.T) {
	/*
		после создания объекта хранения метрик, у него должны быть:
		- пустая карта gauge (len(...) == 0)
		- пустая карта counter (len(...) == 0)
	*/

	storage := NewMemory()
	assert.NotNil(t, storage)
	assert.IsType(t, &MemStorage{}, storage)
	assert.Equal(t, map[string]float64{}, storage.gaugeMap)
	assert.Equal(t, map[string]int64{}, storage.counterMap)
	assert.Len(t, storage.gaugeMap, 0)
	assert.Len(t, storage.counterMap, 0)
}

func TestMemStorage_Metric(t *testing.T) {
	type given struct {
		initGaugeMap   map[string]float64
		initCounterMap map[string]int64
		metric         *models.Metrics
	}

	type want struct {
		metric   *models.Metrics
		hasError bool
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "metric not exists and storage is empty",
			given: given{
				metric: &models.Metrics{},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "metric Gauge not exists and storage is empty",
			given: given{
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "metric Counter not exists and storage is empty",
			given: given{
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "metric Gauge not exists but GaugeMap not empty",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 0,
				},
				metric: &models.Metrics{
					ID:    "GaugeNotExists",
					MType: models.Gauge,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "metric Counter not exists but CounterMap not empty",
			given: given{
				initCounterMap: map[string]int64{
					"Counter": 0,
				},
				metric: &models.Metrics{
					ID:    "CounterNotExists",
					MType: models.Counter,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "metric Gauge exists in GaugeMap",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 0,
				},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(0)),
				},
			},
		},
		{
			name: "metric Counter exists in CounterMap",
			given: given{
				initCounterMap: map[string]int64{
					"Counter": 0,
				},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(0)),
				},
			},
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			storage := &MemStorage{
				gaugeMap:   tt.given.initGaugeMap,
				counterMap: tt.given.initCounterMap,
			}
			cleanupStorage(t, storage)

			metric, err := storage.Metric(ctx, tt.given.metric)
			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.metric, metric)
		})
	}
}

func TestMemStorage_UpdateBatch(t *testing.T) {
	type given struct {
		initGaugeMap   map[string]float64
		initCounterMap map[string]int64
		metric         *models.Metrics
	}

	type want struct {
		metric   *models.Metrics
		hasError bool
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "has error cause empty ID for Gauge",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "",
					MType: models.Gauge,
					Value: new(float64(0)),
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "has error cause empty ID for Counter",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "",
					MType: models.Counter,
					Delta: new(int64(0)),
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "has error cause empty MType for Gauge",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: "",
					Value: new(float64(0)),
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "has error cause empty MType for Counter",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: "",
					Delta: new(int64(0)),
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "has error cause nil value for Gauge",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: nil,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "has error cause nil value for Counter",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: nil,
				},
			},
			want: want{
				hasError: true,
				metric:   nil,
			},
		},
		{
			name: "success updated Gauge on empty map",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(1)),
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(1)),
				},
			},
		},
		{
			name: "success updated Counter on empty map",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(1)),
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(1)),
				},
			},
		},
		{
			name: "success updated Gauge on not empty map",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 1,
				},
				initCounterMap: map[string]int64{},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(2)),
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(2)),
				},
			},
		},
		{
			name: "success updated Counter on not empty map",
			given: given{
				initGaugeMap: map[string]float64{},
				initCounterMap: map[string]int64{
					"Counter": 1,
				},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(2)),
				},
			},
			want: want{
				hasError: false,
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(3)),
				},
			},
		},
	}
	ctx := context.TODO()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				gaugeMap:   tt.given.initGaugeMap,
				counterMap: tt.given.initCounterMap,
			}

			cleanupStorage(t, storage)

			_, err := storage.UpdateBatch(ctx, []models.Metrics{*tt.given.metric})
			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			metric, err := storage.Metric(ctx, tt.given.metric)
			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.metric, metric)
		})
	}
}

func TestMemStorage_MetricList(t *testing.T) {

	type given struct {
		initGaugeMap   map[string]float64
		initCounterMap map[string]int64
		mType          string
	}

	type want struct {
		metrics  []models.Metrics
		hasError bool
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "error cause unknown metric type",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				mType:          "unknownMetricType",
			},
			want: want{
				hasError: true,
				metrics:  nil,
			},
		},
		{
			name: "empty list for Gauge on empty maps",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				mType:          models.Gauge,
			},
			want: want{
				hasError: false,
				metrics:  []models.Metrics{},
			},
		},
		{
			name: "empty list for Counter on empty maps",
			given: given{
				initGaugeMap:   map[string]float64{},
				initCounterMap: map[string]int64{},
				mType:          models.Counter,
			},
			want: want{
				hasError: false,
				metrics:  []models.Metrics{},
			},
		},
		{
			name: "empty list for Gauge on empty map and not empty Counter map",
			given: given{
				initGaugeMap: map[string]float64{},
				initCounterMap: map[string]int64{
					"Counter": 1,
				},
				mType: models.Gauge,
			},
			want: want{
				hasError: false,
				metrics:  []models.Metrics{},
			},
		},
		{
			name: "empty list for Counter on empty map and not empty Gauge map",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 1,
				},
				initCounterMap: map[string]int64{},
				mType:          models.Counter,
			},
			want: want{
				hasError: false,
				metrics:  []models.Metrics{},
			},
		},
		{
			name: "not empty list for Gauge on not empty map and empty Counter map",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 1,
				},
				initCounterMap: map[string]int64{},
				mType:          models.Gauge,
			},
			want: want{
				hasError: false,
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(1)),
					},
				},
			},
		},
		{
			name: "not empty list for Counter on not empty map and empty Gauge map",
			given: given{
				initGaugeMap: map[string]float64{},
				initCounterMap: map[string]int64{
					"Counter": 1,
				},
				mType: models.Counter,
			},
			want: want{
				hasError: false,
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(1)),
					},
				},
			},
		},
		{
			name: "not empty list for Gauge on not empty map and not empty Counter map",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 1,
				},
				initCounterMap: map[string]int64{
					"Counter": 1,
				},
				mType: models.Gauge,
			},
			want: want{
				hasError: false,
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(1)),
					},
				},
			},
		},
		{
			name: "not empty list for Counter on not empty map and not empty Gauge map",
			given: given{
				initGaugeMap: map[string]float64{
					"Gauge": 1,
				},
				initCounterMap: map[string]int64{
					"Counter": 1,
				},
				mType: models.Counter,
			},
			want: want{
				hasError: false,
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(1)),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := &MemStorage{
				gaugeMap:   tt.given.initGaugeMap,
				counterMap: tt.given.initCounterMap,
			}
			cleanupStorage(t, storage)
			ctx := context.TODO()
			metrics, er := storage.MetricList(ctx, tt.given.mType)
			if tt.want.hasError {
				assert.Error(t, er)
			} else {
				assert.NoError(t, er)
			}

			assert.Equal(t, tt.want.metrics, metrics)
		})
	}
}

func TestMemStorage_Ping(t *testing.T) {
	t.Run("ping is nil", func(t *testing.T) {
		memStorage := NewMemory()
		assert.NoError(t, memStorage.Ping(context.TODO()))
	})
}
