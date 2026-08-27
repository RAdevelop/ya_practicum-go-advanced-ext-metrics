package metric

import "context"

/*
Storage - интерфейс по работе с хранилищем метрик
скорее всего этот интерфейс надо разделить на два:
- для counter
- для gauge
*/
//go:generate mockery
type Storage interface {
	GaugeUpdate(name string, value float64) error
	GaugeByName(name string) (float64, error)
	Gauge() map[string]float64

	CounterAdd(name string, value int64) error
	CounterAccumulative() map[string]int64
	CounterAccumulativeByName(name string) (int64, error)
	Ping(context.Context) error
}
