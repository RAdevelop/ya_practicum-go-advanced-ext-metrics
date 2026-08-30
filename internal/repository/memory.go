package repository

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
	baseStorage
}

/*
NewMemory - конструктор для структуры хранения метрик
*/
func NewMemory() *MemStorage {

	return &MemStorage{
		gaugeMap:   make(map[string]float64),
		counterMap: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateBatch(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {

	metrics, err := s.baseStorage.UpdateBatch(ctx, metrics)
	if err != nil {
		return nil, err
	}

	for _, metric := range metrics {

		if metric.ID == "" {
			return nil, fmt.Errorf("%w, metric: %+v", perrors.ErrMetricEmptyID, metric)
		}

		if metric.StrValue() == "" {
			return nil, fmt.Errorf("%w, metric: %+v", perrors.ErrMetricEmptyValue, metric)
		}

		switch metric.MType {
		case models.Counter, models.Gauge:
			s.update(ctx, metric)
		default:
			return nil, fmt.Errorf("%w, metric: %+v", perrors.ErrMetricUnknownType, metric)
		}
	}

	return metrics, nil
}

func (s *MemStorage) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {

	metric, err := s.baseStorage.Metric(ctx, metric)

	if err != nil {
		return nil, err
	}

	switch metric.MType {
	case models.Counter:
		metric, err = s.counterByName(ctx, metric.ID)
	case models.Gauge:
		metric, err = s.gaugeByName(ctx, metric.ID)
	default:
		err = fmt.Errorf("%w: metric: %+v", perrors.ErrMetricNotFound, metric)
	}
	if err != nil {
		return nil, err
	}
	return metric, nil
}

func (s *MemStorage) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {

	metrics, err := s.baseStorage.MetricList(ctx, metricType)
	if err != nil {
		return nil, err
	}

	switch metricType {
	case models.Counter:
		metrics = s.counters(ctx)
	case models.Gauge:
		metrics = s.gauges(ctx)
	default:
		err = fmt.Errorf("%w, metricType: %s", perrors.ErrMetricNotFound, metricType)
	}

	if err != nil {
		metrics = nil
	}

	return metrics, err
}

func (s *MemStorage) gaugeUpdate(ctx context.Context, metric models.Metrics) {
	_ = ctx
	s.gaugeMap[metric.ID] = *metric.Value
}

func (s *MemStorage) gaugeByName(ctx context.Context, mID string) (*models.Metrics, error) {
	_ = ctx
	if value, ok := s.gaugeMap[mID]; ok {
		return &models.Metrics{
			ID:    mID,
			MType: models.Gauge,
			Value: &value,
		}, nil
	}

	return nil, fmt.Errorf("%w: mID = %q", perrors.ErrMetricNotFound, mID)
}

func (s *MemStorage) gauges(ctx context.Context) []models.Metrics {
	_ = ctx

	metrics := make([]models.Metrics, 0, len(s.gaugeMap))
	for mID, value := range s.gaugeMap {
		metrics = append(metrics, models.Metrics{
			ID:    mID,
			MType: models.Gauge,
			Value: &value,
		})
	}

	return metrics
}

func (s *MemStorage) update(ctx context.Context, metric models.Metrics) {
	_ = ctx

	switch metric.MType {
	case models.Counter:

		s.counterMap[metric.ID] += *metric.Delta
	case models.Gauge:
		s.gaugeMap[metric.ID] = *metric.Value
	}
}

func (s *MemStorage) counters(ctx context.Context) []models.Metrics {
	_ = ctx
	metrics := make([]models.Metrics, 0, len(s.counterMap))
	for mID, value := range s.counterMap {
		metrics = append(metrics, models.Metrics{
			ID:    mID,
			MType: models.Counter,
			Delta: &value,
		})
	}

	return metrics
}

func (s *MemStorage) counterByName(ctx context.Context, mID string) (*models.Metrics, error) {
	_ = ctx
	if value, ok := s.counterMap[mID]; ok {
		return &models.Metrics{
			ID:    mID,
			MType: models.Counter,
			Delta: &value,
		}, nil
	}
	return nil, fmt.Errorf("%w: mID: %q", perrors.ErrMetricNotFound, mID)
}
