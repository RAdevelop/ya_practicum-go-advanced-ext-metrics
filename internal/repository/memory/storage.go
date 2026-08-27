package memory

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFoundName = errors.New("metric not found by name")

// MemStorage - хранилище метрик в памяти
type MemStorage struct {
	gauge               map[string]float64
	counter             map[string][]int64
	counterAccumulative map[string]int64
}

/*
NewStorage - конструктор для структуры хранения метрик
*/
func NewStorage() *MemStorage {

	return &MemStorage{
		gauge:               make(map[string]float64),
		counter:             make(map[string][]int64),
		counterAccumulative: make(map[string]int64),
	}
}

func (ms *MemStorage) GaugeUpdate(name string, value float64) {
	ms.gauge[name] = value
}

func (ms *MemStorage) GaugeByName(name string) (float64, error) {
	if value, ok := ms.gauge[name]; ok {
		return value, nil
	}

	return 0, fmt.Errorf("%w: name = %q", ErrNotFoundName, name)
}

func (ms *MemStorage) Gauge() map[string]float64 {
	return ms.gauge
}

func (ms *MemStorage) GaugeSize() int {
	return len(ms.gauge)
}

func (ms *MemStorage) CounterAdd(name string, value int64) {
	if _, ok := ms.counter[name]; !ok {
		ms.counter[name] = make([]int64, 0)
	}

	ms.counter[name] = append(ms.counter[name], value)
	ms.counterAccumulate(name, value)
}

func (ms *MemStorage) CounterByName(name string) ([]int64, error) {
	if value, ok := ms.counter[name]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("%w: name = %q", ErrNotFoundName, name)
}

func (ms *MemStorage) CounterAccumulative() map[string]int64 {
	return ms.counterAccumulative
}

func (ms *MemStorage) CounterSize() int {
	return len(ms.counter)
}

func (ms *MemStorage) CounterSizeByName(name string) int {

	if _, ok := ms.counter[name]; !ok {
		return 0
	}

	return len(ms.counter[name])
}

func (ms *MemStorage) CounterAccumulativeByName(name string) (int64, error) {

	if value, ok := ms.counterAccumulative[name]; ok {
		return value, nil
	}
	return 0, fmt.Errorf("%w: name = %q", ErrNotFoundName, name)
}

func (ms *MemStorage) counterAccumulate(name string, value int64) {
	ms.counterAccumulative[name] += value
}

func (ms *MemStorage) Ping(context.Context) error {
	return nil
}
