package repository

import (
	"context"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/perrors"
)

// baseStorage - родительский объект для реализации хранилищ.
// Приватный, чтобы от него не создавали реальные объекты
type baseStorage struct {
}

var availableMetricTypes = map[string]bool{
	models.Gauge:   true,
	models.Counter: true,
}

// AddValuesForDeduplicateMetrics - если в `metrics` есть несколько "Counter" метрик с одинаковым ID, то их значения суммируются.
// А если несколько "Gauge" метрик с одинаковым ID, то остается последняя.
//
// Работает in place: результат разделяет память с входным слайсом.
func (s *baseStorage) AddValuesForDeduplicateMetrics(metrics []models.Metrics) []models.Metrics {
	if len(metrics) < 2 {
		return metrics
	}

	type metricKey struct {
		ID    string
		MType string
	}
	// значение карты — индекс первого вхождения, а не копия структуры:
	// экономит и память карты, и аллокации
	pos := make(map[metricKey]int, len(metrics))

	// фаза 1: сворачиваем дубликаты в первые вхождения
	for i, m := range metrics {
		if !s.IsTypeAvailable(m.MType) {
			continue
		}
		key := metricKey{ID: m.ID, MType: m.MType}
		j, ok := pos[key]
		if !ok {
			pos[key] = i
			continue
		}

		switch m.MType {
		case models.Gauge:
			// берем последнее значение, так как Gauge обновляется, а не суммируется
			metrics[j] = m
		case models.Counter:
			*metrics[j].Delta += *m.Delta
		}
	}

	// фаза 2: без аллокации — переиспользуем входной слайс
	out := metrics[:0]
	for i, m := range metrics {
		key := metricKey{ID: m.ID, MType: m.MType}
		if j, ok := pos[key]; ok && j == i {
			out = append(out, m)
		}
	}

	return out
}

func (s *baseStorage) IsTypeAvailable(mType string) bool {
	if isAvailable, ok := availableMetricTypes[mType]; ok {
		return isAvailable
	}

	return false
}

func (s *baseStorage) Ping(context.Context) error {
	return nil
}

func (s *baseStorage) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
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

func (s *baseStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	metrics = s.AddValuesForDeduplicateMetrics(metrics)

	if len(metrics) == 0 {
		return nil, fmt.Errorf("%w", perrors.ErrMetricListEmpty)
	}
	return metrics, nil
}

func (s *baseStorage) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !s.IsTypeAvailable(metricType) {
		return nil, fmt.Errorf("%w, metricType: %s", perrors.ErrMetricUnknownType, metricType)
	}

	return []models.Metrics{}, nil
}
