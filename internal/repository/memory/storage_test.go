package memory

import (
	"context"
	"testing"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMemStorage_NewMemStorage(t *testing.T) {
	/*
		после создания объекта хранения метрик, у него должны быть:
		- пустая карта gauge (len(...) == 0)
		- пустая карта counter (len(...) == 0)
	*/

	storage := NewStorage()
	assert.NotNil(t, storage)
	assert.IsType(t, &MemStorage{}, storage)
	assert.Equal(t, map[string]float64{}, storage.gauge)
	assert.Equal(t, map[string][]int64{}, storage.counter)
	assert.Len(t, storage.gauge, 0)
	assert.Len(t, storage.counter, 0)

}

func TestMemStorage_validateName(t *testing.T) {

}

func TestMemStorage_CounterByNameAccumulative(t *testing.T) {

	type got struct {
		counters          []int64
		counterMetricName string
	}

	type want struct {
		counterAccumulative *models.Metrics
		err                 error
	}

	tests := []struct {
		name string
		got  got
		want want
	}{
		{
			name: "counter accumulative empty",
			got: got{
				counters:          []int64{},
				counterMetricName: "counter",
			},
			want: want{
				counterAccumulative: nil,
				err:                 ErrNotFoundName,
			},
		},
		{
			name: "counter accumulative none empty",
			got: got{
				counters: []int64{
					0,
				},
				counterMetricName: "counter",
			},
			want: want{
				counterAccumulative: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(0)),
				},
				err: nil,
			},
		},
		{
			name: "counter accumulative none empty",
			got: got{
				counters: []int64{
					0, 1, 2, 3,
				},
				counterMetricName: "counter",
			},
			want: want{
				counterAccumulative: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(6)),
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := NewStorage()

			for _, counter := range tt.got.counters {
				err := memStorage.CounterAdd(context.TODO(), tt.got.counterMetricName, counter)
				assert.NoError(t, err)
			}

			counterAccumulative, err := memStorage.CounterAccumulativeByName(context.TODO(), tt.got.counterMetricName)
			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want.counterAccumulative, counterAccumulative)
		})
	}
}

func TestMemStorage_CounterAdd(t *testing.T) {

	type got struct {
		memStorage         MemStorage
		counterMetricName  string
		counterMetricValue int64
	}

	type want struct {
		lenCounterMetricNameAfterAdd int
		hasError                     bool
	}

	tests := []struct {
		name string
		got  got
		want want
	}{
		{
			name: "add to empty counter map",
			got: got{
				memStorage: MemStorage{
					counter:             map[string][]int64{},
					gauge:               map[string]float64{},
					counterAccumulative: map[string]int64{},
				},
				counterMetricName:  "counter",
				counterMetricValue: 123,
			},
			want: want{
				lenCounterMetricNameAfterAdd: 1,
				hasError:                     false,
			},
		},
		{
			name: "add to none empty counter map",
			got: got{
				memStorage: MemStorage{
					counter: map[string][]int64{
						"counter": {1, 2, 3},
					},
					gauge:               map[string]float64{},
					counterAccumulative: map[string]int64{},
				},
				counterMetricName:  "counter",
				counterMetricValue: 123,
			},
			want: want{
				lenCounterMetricNameAfterAdd: 4,
				hasError:                     false,
			},
		},
		{
			name: "add to none empty counter map with another metric",
			got: got{
				memStorage: MemStorage{
					counter: map[string][]int64{
						"counter": {1, 2, 3},
					},
					gauge:               map[string]float64{},
					counterAccumulative: map[string]int64{},
				},
				counterMetricName:  "anotherCounter",
				counterMetricValue: 123,
			},
			want: want{
				lenCounterMetricNameAfterAdd: 1,
				hasError:                     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			memStorage := &tt.got.memStorage

			err := memStorage.CounterAdd(context.TODO(), tt.got.counterMetricName, tt.got.counterMetricValue)
			assert.NoError(t, err)
		})
	}
}

func TestMemStorage_GaugeUpdate(t *testing.T) {
	tests := []struct {
		name                string
		gaugeMapInit        map[string]float64
		gaugeMapAfterUpdate []models.Metrics
		gaugeMetricName     string
		gaugeMetricValue    float64
	}{
		{
			name:         "update to empty gauge map with valid metric name",
			gaugeMapInit: map[string]float64{},
			gaugeMapAfterUpdate: []models.Metrics{
				{
					ID:    "validMetricName",
					MType: models.Gauge,
					Value: new(float64(1)),
				},
			},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: 1,
		},
		{
			name: "update exists metric",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1,
			},
			gaugeMapAfterUpdate: []models.Metrics{
				{
					ID:    "validMetricName",
					MType: models.Gauge,
					Value: new(float64(2)),
				},
			},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: 2,
		},
		{
			name: "add new metric",
			gaugeMapInit: map[string]float64{
				"validMetricName": 0,
			},
			gaugeMapAfterUpdate: []models.Metrics{
				{
					ID:    "validMetricName",
					MType: models.Gauge,
					Value: new(float64(0)),
				},
				{
					ID:    "validMetricNameNew",
					MType: models.Gauge,
					Value: new(float64(2)),
				},
			},
			gaugeMetricName:  "validMetricNameNew",
			gaugeMetricValue: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := &MemStorage{
				gauge: tt.gaugeMapInit,
			}

			err := memStorage.GaugeUpdate(context.TODO(), tt.gaugeMetricName, tt.gaugeMetricValue)
			assert.NoError(t, err)
			assert.Equal(t, memStorage.gauge, tt.gaugeMapInit)

			gauge, err := memStorage.Gauge(context.TODO())
			assert.NoError(t, err)
			assert.ElementsMatch(t, tt.gaugeMapAfterUpdate, gauge)
		})
	}
}

func TestMemStorage_GaugeByName(t *testing.T) {
	tests := []struct {
		name             string
		gaugeMapInit     map[string]float64
		gaugeMetricName  string
		gaugeMetricValue *models.Metrics
		err              error
	}{
		{
			name:             "get gauge by name valid metric name but not found",
			gaugeMapInit:     map[string]float64{},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: nil,
			err:              ErrNotFoundName,
		},
		{
			name: "success get gauge by name valid metric with value 1",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1,
			},
			gaugeMetricName: "validMetricName",
			gaugeMetricValue: &models.Metrics{
				ID:    "validMetricName",
				MType: models.Gauge,
				Value: new(float64(1)),
			},
			err: nil,
		},
		{
			name: "success get gauge by name valid metric with value 1.2",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1.2,
			},
			gaugeMetricName: "validMetricName",
			gaugeMetricValue: &models.Metrics{
				ID:    "validMetricName",
				MType: models.Gauge,
				Value: new(1.2),
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := &MemStorage{
				gauge: tt.gaugeMapInit,
			}

			metricValue, err := memStorage.GaugeByName(context.TODO(), tt.gaugeMetricName)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err, "GaugeByName() should return an error for name: %s", tt.gaugeMetricName)
			} else {
				assert.Equal(t, tt.gaugeMetricValue, metricValue)
			}
		})
	}
}

func TestMemStorage_Ping(t *testing.T) {
	t.Run("ping is nil", func(t *testing.T) {
		memStorage := NewStorage()
		assert.NoError(t, memStorage.Ping(context.TODO()))
	})
}
