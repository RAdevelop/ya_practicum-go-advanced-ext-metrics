package metric

import (
	"context"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

// Service - сервис для работы с метриками
type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (ms *Service) GaugeUpdate(ctx context.Context, name string, value float64) error {
	return ms.storage.GaugeUpdate(ctx, name, value)
}

func (ms *Service) CounterAdd(ctx context.Context, name string, value int64) error {
	return ms.storage.CounterAdd(ctx, name, value)
}

func (ms *Service) CounterByNameAccumulative(ctx context.Context, name string) (*models.Metrics, error) {
	return ms.storage.CounterAccumulativeByName(ctx, name)
}

func (ms *Service) GaugeByName(ctx context.Context, name string) (*models.Metrics, error) {
	return ms.storage.GaugeByName(ctx, name)
}

func (ms *Service) Gauge(ctx context.Context) ([]models.Metrics, error) {
	return ms.storage.Gauge(ctx)
}

func (ms *Service) CounterAccumulative(ctx context.Context) ([]models.Metrics, error) {
	return ms.storage.CounterAccumulative(ctx)
}

func (ms *Service) Ping(ctx context.Context) error {
	return ms.storage.Ping(ctx)
}
