package service

import (
	"context"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
)

type MetricManagementAble interface {
	MetricUpdate(context.Context, *models.Metrics) error
	MetricValue(context.Context, string, string) (*models.Metrics, error)
	MetricList(context.Context, string) map[string]models.Metrics
	MetricSnapshotLoad(context.Context) error
	MetricSnapshotSave(context.Context) error
	StoragePing(context.Context) error
}

// Manager - предоставляет интерфейс (фасад) для работы с метриками
type Manager struct {
	metricService  *metric.Service
	metricSnapshot snapshot.Able
}

func NewManager(metricService *metric.Service, metricSnapshot snapshot.Able) *Manager {
	return &Manager{
		metricService:  metricService,
		metricSnapshot: metricSnapshot,
	}
}

func (manager *Manager) MetricUpdate(ctx context.Context, metric *models.Metrics) error {
	if metric.MType == models.Counter && metric.Delta != nil {
		return manager.metricService.CounterAdd(ctx, metric.ID, *metric.Delta)
	} else if metric.MType == models.Gauge && metric.Value != nil {
		return manager.metricService.GaugeUpdate(ctx, metric.ID, *metric.Value)
	}
	//TODO вернуть ошибку "unknown metric'?
	return nil
}

func (manager *Manager) MetricValue(ctx context.Context, metricType string, metricID string) (*models.Metrics, error) {
	var modelMetric *models.Metrics

	if metricType == models.Counter {
		value, err := manager.metricService.CounterByNameAccumulative(ctx, metricID)
		if err != nil {
			return nil, err
		}
		modelMetric = &models.Metrics{
			MType: models.Counter,
			ID:    metricID,
			Delta: &value,
		}
	} else if metricType == models.Gauge {
		value, err := manager.metricService.GaugeByName(ctx, metricID)
		if err != nil {
			return nil, err
		}
		modelMetric = &models.Metrics{
			MType: models.Gauge,
			ID:    metricID,
			Value: &value,
		}
	}

	return modelMetric, nil
}

func (manager *Manager) MetricList(ctx context.Context, metricType string) map[string]models.Metrics {

	metricsCounter := manager.metricService.CounterAccumulative(ctx)
	metricsGauge := manager.metricService.Gauge(ctx)

	metrics := make(map[string]models.Metrics, len(metricsCounter)+len(metricsGauge))

	if metricType == models.Counter {
		for name, value := range metricsCounter {
			metrics[name] = models.Metrics{
				MType: models.Counter,
				ID:    name,
				Delta: &value,
			}
		}
	} else if metricType == models.Gauge {
		for name, value := range metricsGauge {
			metrics[name] = models.Metrics{
				MType: models.Gauge,
				ID:    name,
				Value: &value,
			}
		}
	}

	return metrics
}

func (manager *Manager) MetricSnapshotLoad(ctx context.Context) error {
	return manager.metricSnapshot.Load(ctx)
}
func (manager *Manager) MetricSnapshotSave(ctx context.Context) error {
	return manager.metricSnapshot.Save(ctx)
}

func (manager *Manager) StoragePing(ctx context.Context) error {
	return manager.metricService.Ping(ctx)
}
