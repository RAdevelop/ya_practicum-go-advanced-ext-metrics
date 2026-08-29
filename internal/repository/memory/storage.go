package memory

import (
	"context"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/perrors"
)

// MemStorage - хранилище метрик в памяти
type MemStorage struct {
	gaugeMap   map[string]float64
	counterMap map[string]int64
}

/*
NewStorage - конструктор для структуры хранения метрик
*/
func NewStorage() *MemStorage {

	return &MemStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

func (ms *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	_ = ctx

	for _, metric := range metrics {

		if metric.ID == "" {
			return fmt.Errorf("%w: %+v", perrors.ErrMetricEmptyID, metric)
		}

		if metric.StrValue() == "" {
			return fmt.Errorf("%w: %+v", perrors.ErrMetricEmptyValue, metric)
		}

		switch metric.MType {
		case models.Counter, models.Gauge:
			ms.update(ctx, metric)
		default:
			return fmt.Errorf("%w: %+v", perrors.ErrMetricUnknownType, metric)
		}
	}

	return nil
}

func (ms *MemStorage) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {

	var err error

	switch metric.MType {
	case models.Counter:
		metric, err = ms.counterByName(ctx, metric.ID)
	case models.Gauge:
		metric, err = ms.gaugeByName(ctx, metric.ID)
	default:
		err = fmt.Errorf("%w: metric: %+v", perrors.ErrMetricNotFound, metric)
	}
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (ms *MemStorage) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	var metrics []models.Metrics
	var err error

	switch metricType {
	case models.Counter:
		metrics = ms.counters(ctx)
	case models.Gauge:
		metrics = ms.gauges(ctx)
	default:
		err = fmt.Errorf("%w, metricType: %s", perrors.ErrMetricUnknownType, metricType)
	}

	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (ms *MemStorage) gaugeUpdate(ctx context.Context, metric models.Metrics) {
	_ = ctx
	ms.gaugeMap[metric.ID] = *metric.Value
}

func (ms *MemStorage) gaugeByName(ctx context.Context, mID string) (*models.Metrics, error) {
	_ = ctx
	if value, ok := ms.gaugeMap[mID]; ok {
		return &models.Metrics{
			ID:    mID,
			MType: models.Gauge,
			Value: &value,
		}, nil
	}

	return nil, fmt.Errorf("%w: mID = %q", perrors.ErrMetricNotFound, mID)
}

func (ms *MemStorage) gauges(ctx context.Context) []models.Metrics {
	_ = ctx

	metrics := make([]models.Metrics, 0, len(ms.gaugeMap))
	for mID, value := range ms.gaugeMap {
		metrics = append(metrics, models.Metrics{
			ID:    mID,
			MType: models.Gauge,
			Value: &value,
		})
	}

	return metrics
}

func (ms *MemStorage) update(ctx context.Context, metric models.Metrics) {
	_ = ctx

	switch metric.MType {
	case models.Counter:

		ms.counterMap[metric.ID] += *metric.Delta
	case models.Gauge:
		ms.gaugeMap[metric.ID] = *metric.Value
	}
}

func (ms *MemStorage) counters(ctx context.Context) []models.Metrics {
	_ = ctx
	metrics := make([]models.Metrics, 0, len(ms.counterMap))
	for mID, value := range ms.counterMap {
		metrics = append(metrics, models.Metrics{
			ID:    mID,
			MType: models.Counter,
			Delta: &value,
		})
	}

	return metrics
}

func (ms *MemStorage) counterByName(ctx context.Context, mID string) (*models.Metrics, error) {
	_ = ctx
	if value, ok := ms.counterMap[mID]; ok {
		return &models.Metrics{
			ID:    mID,
			MType: models.Counter,
			Delta: &value,
		}, nil
	}
	return nil, fmt.Errorf("%w: mID: %q", perrors.ErrMetricNotFound, mID)
}

func (ms *MemStorage) Ping(context.Context) error {
	return nil
}
