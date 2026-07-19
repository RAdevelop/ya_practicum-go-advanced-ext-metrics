package service

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

	memStorage := NewMemStorage()
	assert.NotNil(t, memStorage)
	assert.IsType(t, &MemStorage{}, memStorage)
	assert.Equal(t, map[string]float64{}, memStorage.gauge)
	assert.Equal(t, map[string][]int64{}, memStorage.counter)
	assert.Len(t, memStorage.gauge, 0)
	assert.Len(t, memStorage.counter, 0)

}

func TestMemStorage_validateName(t *testing.T) {

	tests := []struct {
		name   string
		hasErr bool
		names  []string
	}{
		{
			name:   "valid name",
			hasErr: false,
			names: []string{
				"name",
				"nameValid",
				"nameValid123",
				"name123Valid",
				"nameValid0",
				"nameValid-",
				"nameValid_",
				"nameValid_-",
				"nameValid_1-2_3",
				"name-Valid_1-2_3",
				"name-1_2_0_Valid_1-2_3",
				"name-1.2_0.Valid_1-2_3.",
			},
		},
		{
			name:   "invalid name",
			hasErr: true,
			names: []string{
				"",
				"0",
				"1",
				"123",
				"1234",
				"1name123Valid",
				"-nameInValid",
				"-nameInValid-",
				".nameInValid",
				".nameInValid.",
				"_nameInValid",
				"_nameInValid_",
				"-1nameInValid",
				"-1nameInValid-",
				".1nameInValid",
				".1nameInValid.",
				"_1nameInValid",
				"_1nameInValid_",
				"-",
				".",
				"-",
				"-1",
				".2",
				"-3",
				"----",
				"____",
				"....",
				"-123",
				"_123",
				".123",
			},
		},
	}
	memStorage := NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, name := range tt.names {
				err := memStorage.validateName(name)
				if tt.hasErr {
					assert.ErrorIs(t, err, ErrNameInvalid, "validateName should return an error for name: %s", name)
				} else {
					assert.NoError(t, err, "validateName should not return an error for name: %s", name)
				}
			}
		})
	}
}

func TestMemStorage_CounterAdd(t *testing.T) {

	tests := []struct {
		name                         string
		counterMapInit               map[string][]int64
		counterMetricName            string
		counterMetricValue           int64
		lenCounterAfterAdd           int
		lenCounterMetricNameAfterAdd int
		hasError                     bool
	}{
		{
			name:                         "add to empty counter map",
			counterMapInit:               map[string][]int64{},
			counterMetricName:            "counter",
			counterMetricValue:           123,
			lenCounterAfterAdd:           1,
			lenCounterMetricNameAfterAdd: 1,
			hasError:                     false,
		},
		{
			name: "add to none empty counter map",
			counterMapInit: map[string][]int64{
				"counter": {1, 2, 3},
			},
			counterMetricName:            "counter",
			counterMetricValue:           123,
			lenCounterAfterAdd:           1,
			lenCounterMetricNameAfterAdd: 4,
			hasError:                     false,
		},
		{
			name: "add to none empty counter map with another metric",
			counterMapInit: map[string][]int64{
				"counter": {1, 2, 3},
			},
			counterMetricName:            "anotherCounter",
			counterMetricValue:           123,
			lenCounterAfterAdd:           2,
			lenCounterMetricNameAfterAdd: 1,
			hasError:                     false,
		},
		{
			name:                         "add invalid metric name",
			counterMapInit:               map[string][]int64{},
			counterMetricName:            "invalid metric name",
			counterMetricValue:           123,
			lenCounterAfterAdd:           0,
			lenCounterMetricNameAfterAdd: 0,
			hasError:                     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			memStorage := &MemStorage{
				counter: tt.counterMapInit,
			}

			err := memStorage.CounterAdd(tt.counterMetricName, tt.counterMetricValue)

			if tt.hasError {
				assert.ErrorIs(t, err, ErrNameInvalid, "counter should return an error for name: %v", ErrNameInvalid)
				return
			}

			assert.NoError(t, err, "CounterAdd should not return an error")
			assert.Equal(t, memStorage.CounterSize(), tt.lenCounterAfterAdd)
			assert.Equal(t, memStorage.CounterSizeByName(tt.counterMetricName), tt.lenCounterMetricNameAfterAdd)
			assert.Equal(t, tt.counterMapInit, memStorage.Counter())
		})
	}
}
