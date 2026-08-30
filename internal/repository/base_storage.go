package repository

import (
	"context"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/perrors"
)

type BaseStorage struct {
}

var availableMetricTypes = map[string]bool{
	models.Gauge:   true,
	models.Counter: true,
}

// AddValuesForDeduplicateMetrics - если в `metrics` если несколько "Counter" метрик с одинаковым ID, то их значения суммируются.
// А если несколько "Gauge" метрик с одинаковым ID, то остается последняя.
func (s *BaseStorage) AddValuesForDeduplicateMetrics(metrics []models.Metrics) []models.Metrics {

	if len(metrics) == 0 {
		return metrics
	}

	cleanMetrics := make(map[string]models.Metrics, len(metrics))

	for _, m := range metrics {
		key := fmt.Sprintf("%s|%s", m.ID, m.MType)

		// не будем сохранять метрики с неизвестным типом
		if !s.IsTypeAvailable(m.MType) {
			continue
		}

		if _, ok := cleanMetrics[key]; !ok {
			cleanMetrics[key] = m
			continue
		}

		switch m.MType {
		case models.Gauge:
			// берем последнее значение, так как Gauge обновляется, а не суммируется
			cleanMetrics[key] = m
		case models.Counter:
			*cleanMetrics[key].Delta += *m.Delta
		}
	}

	if len(cleanMetrics) == 0 {
		metrics = nil
		return nil
	}

	summarizedMetrics := make([]models.Metrics, 0, len(cleanMetrics))

	for _, m := range cleanMetrics {
		summarizedMetrics = append(summarizedMetrics, m)
	}
	cleanMetrics = nil
	return summarizedMetrics
}

func (s *BaseStorage) IsTypeAvailable(mType string) bool {
	if isAvailable, ok := availableMetricTypes[mType]; ok {
		return isAvailable
	}

	return false
}

func (s *BaseStorage) Ping(context.Context) error {
	return nil
}

func (s *BaseStorage) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if metric == nil {
		return nil, fmt.Errorf("%w", perrors.ErrMetricIsNil)
	}

	if !s.IsTypeAvailable(metric.MType) {
		return nil, fmt.Errorf("%w, metric: %+v", perrors.ErrMetricUnknownType, metric)
	}

	if metric.ID == "" {
		return nil, fmt.Errorf("%w", perrors.ErrMetricEmptyID)
	}

	return metric, nil
}

func (s *BaseStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	metrics = s.AddValuesForDeduplicateMetrics(metrics)

	if len(metrics) == 0 {
		return nil, fmt.Errorf("%w", perrors.ErrMetricListEmpty)
	}
	return metrics, nil
}

func (s *BaseStorage) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	if s.IsTypeAvailable(metricType) {
		return nil, fmt.Errorf("%w, metricType: %s", perrors.ErrMetricUnknownType, metricType)
	}

	return []models.Metrics{}, nil
}
