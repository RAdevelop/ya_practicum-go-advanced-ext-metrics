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

func (ms *MemStorage) GaugeUpdate(ctx context.Context, name string, value float64) error {
	_ = ctx
	ms.gauge[name] = value
	return nil
}

func (ms *MemStorage) GaugeByName(ctx context.Context, name string) (float64, error) {
	_ = ctx
	if value, ok := ms.gauge[name]; ok {
		return value, nil
	}

	return 0, fmt.Errorf("%w: name = %q", ErrNotFoundName, name)
}

func (ms *MemStorage) Gauge(ctx context.Context) map[string]float64 {
	_ = ctx
	return ms.gauge
}

func (ms *MemStorage) CounterAdd(ctx context.Context, name string, value int64) error {
	_ = ctx
	if _, ok := ms.counter[name]; !ok {
		ms.counter[name] = make([]int64, 0)
	}

	ms.counter[name] = append(ms.counter[name], value)
	ms.counterAccumulate(name, value)
	return nil
}

func (ms *MemStorage) CounterByName(ctx context.Context, name string) ([]int64, error) {
	_ = ctx
	if value, ok := ms.counter[name]; ok {
		return value, nil
	}

	return nil, fmt.Errorf("%w: name = %q", ErrNotFoundName, name)
}

func (ms *MemStorage) CounterAccumulative(ctx context.Context) map[string]int64 {
	_ = ctx
	return ms.counterAccumulative
}

func (ms *MemStorage) CounterSizeByName(ctx context.Context, name string) int {
	_ = ctx
	if _, ok := ms.counter[name]; !ok {
		return 0
	}

	return len(ms.counter[name])
}

func (ms *MemStorage) CounterAccumulativeByName(ctx context.Context, name string) (int64, error) {
	_ = ctx
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
