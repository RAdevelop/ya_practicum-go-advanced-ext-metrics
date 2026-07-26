package memory

import (
	"testing"

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

func TestMemStorage_CounterByName(t *testing.T) {

	tests := []struct {
		name              string
		counterMapInit    map[string][]int64
		counterMetricName string
		err               error
	}{
		{
			name:              "error is ErrNotFoundName",
			counterMapInit:    map[string][]int64{},
			counterMetricName: "counter",
			err:               ErrNotFoundName,
		},
		{
			name: "metric found by name",
			counterMapInit: map[string][]int64{
				"counter": {1, 2, 3},
			},
			counterMetricName: "counter",
			err:               nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := &MemStorage{
				counter: tt.counterMapInit,
			}

			counterMetric, err := memStorage.CounterByName(tt.counterMetricName)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err, "CounterByName() should return an error for name: %s", tt.counterMetricName)
				assert.Nil(t, counterMetric)
			} else {
				assert.NoError(t, err, "CounterByName() should not return an error for name: %s", tt.counterMetricName)
				assert.Equal(t, tt.counterMapInit[tt.counterMetricName], counterMetric)
			}
		})
	}

}

func TestMemStorage_CounterByNameAccumulative(t *testing.T) {

	type got struct {
		counters          []int64
		counterMetricName string
	}

	type want struct {
		counterAccumulative int64
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
				counterAccumulative: 0,
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
				counterAccumulative: 0,
				err:                 nil,
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
				counterAccumulative: 6,
				err:                 nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := NewStorage()

			for _, counter := range tt.got.counters {
				memStorage.CounterAdd(tt.got.counterMetricName, counter)
			}

			counterAccumulative, err := memStorage.CounterAccumulativeByName(tt.got.counterMetricName)
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
		lenCounterAfterAdd           int
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
				lenCounterAfterAdd:           1,
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
				lenCounterAfterAdd:           1,
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
				lenCounterAfterAdd:           2,
				lenCounterMetricNameAfterAdd: 1,
				hasError:                     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			memStorage := &tt.got.memStorage

			memStorage.CounterAdd(tt.got.counterMetricName, tt.got.counterMetricValue)
			assert.Equal(t, memStorage.CounterSize(), tt.want.lenCounterAfterAdd)
			assert.Equal(t, memStorage.CounterSizeByName(tt.got.counterMetricName), tt.want.lenCounterMetricNameAfterAdd)
		})
	}
}

func TestMemStorage_GaugeUpdate(t *testing.T) {
	tests := []struct {
		name                 string
		gaugeMapInit         map[string]float64
		gaugeMapAfterUpdate  map[string]float64
		gaugeMetricName      string
		gaugeMetricValue     float64
		gaugeSizeAfterUpdate int
	}{
		{
			name:         "update to empty gauge map with valid metric name",
			gaugeMapInit: map[string]float64{},
			gaugeMapAfterUpdate: map[string]float64{
				"validMetricName": 1,
			},
			gaugeMetricName:      "validMetricName",
			gaugeMetricValue:     1,
			gaugeSizeAfterUpdate: 1,
		},
		{
			name: "update exists metric",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1,
			},
			gaugeMapAfterUpdate: map[string]float64{
				"validMetricName": 2,
			},
			gaugeMetricName:      "validMetricName",
			gaugeMetricValue:     2,
			gaugeSizeAfterUpdate: 1,
		},
		{
			name: "add new metric",
			gaugeMapInit: map[string]float64{
				"validMetricName": 0,
			},
			gaugeMapAfterUpdate: map[string]float64{
				"validMetricName":    0,
				"validMetricNameNew": 2,
			},
			gaugeMetricName:      "validMetricNameNew",
			gaugeMetricValue:     2,
			gaugeSizeAfterUpdate: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := &MemStorage{
				gauge: tt.gaugeMapInit,
			}

			memStorage.GaugeUpdate(tt.gaugeMetricName, tt.gaugeMetricValue)

			assert.Equal(t, memStorage.gauge, tt.gaugeMapInit)
			assert.Equal(t, memStorage.GaugeSize(), tt.gaugeSizeAfterUpdate)
			assert.Equal(t, memStorage.Gauge(), tt.gaugeMapAfterUpdate)
		})
	}
}

func TestMemStorage_GaugeByName(t *testing.T) {
	tests := []struct {
		name             string
		gaugeMapInit     map[string]float64
		gaugeMetricName  string
		gaugeMetricValue float64
		err              error
	}{
		{
			name:             "get gauge by name valid metric name but not found",
			gaugeMapInit:     map[string]float64{},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: 0,
			err:              ErrNotFoundName,
		},
		{
			name: "success get gauge by name valid metric with value 1",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1,
			},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: 1,
			err:              nil,
		},
		{
			name: "success get gauge by name valid metric with value 1.2",
			gaugeMapInit: map[string]float64{
				"validMetricName": 1.2,
			},
			gaugeMetricName:  "validMetricName",
			gaugeMetricValue: 1.2,
			err:              nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			memStorage := &MemStorage{
				gauge: tt.gaugeMapInit,
			}

			metricValue, err := memStorage.GaugeByName(tt.gaugeMetricName)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err, "GaugeByName() should return an error for name: %s", tt.gaugeMetricName)
			} else {
				assert.NoError(t, err, "GaugeByName() should not return an error for name: %s", tt.gaugeMetricName)
				assert.Equal(t, memStorage.Gauge(), tt.gaugeMapInit)
				assert.Equal(t, tt.gaugeMetricValue, metricValue)
			}
		})
	}
}
