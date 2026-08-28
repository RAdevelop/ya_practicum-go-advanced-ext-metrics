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
	GaugeUpdate(ctx context.Context, name string, value float64) error
	GaugeByName(ctx context.Context, name string) (float64, error)
	Gauge(ctx context.Context) map[string]float64

	CounterAdd(ctx context.Context, name string, value int64) error
	CounterAccumulative(ctx context.Context) map[string]int64
	CounterAccumulativeByName(ctx context.Context, name string) (int64, error)
	Ping(context.Context) error
}
