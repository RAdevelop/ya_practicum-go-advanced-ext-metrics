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
	UpdateBatch(ctx context.Context, metrics []models.Metrics) error
	Metric(ctx context.Context, metrics *models.Metrics) (*models.Metrics, error)
	MetricList(ctx context.Context, metricType string) ([]models.Metrics, error)
	Ping(context.Context) error
}
