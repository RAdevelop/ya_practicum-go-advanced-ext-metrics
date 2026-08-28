package metric

import (
	"context"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

/*
Storage - интерфейс по работе с хранилищем метрик
скорее всего этот интерфейс надо разделить на два:
- для counter
- для gauge
*/
//go:generate mockery
type Storage interface {
	GaugeUpdate(ctx context.Context, name string, value float64) error
	GaugeByName(ctx context.Context, name string) (*models.Metrics, error)
	Gauge(ctx context.Context) ([]models.Metrics, error)

	CounterAdd(ctx context.Context, name string, value int64) error
	CounterAccumulative(ctx context.Context) ([]models.Metrics, error)
	CounterAccumulativeByName(ctx context.Context, name string) (*models.Metrics, error)
	Ping(context.Context) error
}
