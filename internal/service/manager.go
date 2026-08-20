package service

import (
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
)

type MetricManagementAble interface {
	MetricUpdate(*models.Metrics)
	MetricValue(string, string) (*models.Metrics, error)
	MetricList(string) map[string]models.Metrics
	MetricSnapshotLoad() error
	MetricSnapshotSave() error
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

func (manager *Manager) MetricUpdate(metric *models.Metrics) {
	if metric.MType == models.Counter {
		if metric.Delta != nil {
			manager.metricService.CounterAdd(metric.ID, *metric.Delta)
		}
	} else {
		if metric.Value != nil {
			manager.metricService.GaugeUpdate(metric.ID, *metric.Value)
		}
	}
}

func (manager *Manager) MetricValue(metricType string, metricID string) (*models.Metrics, error) {
	var modelMetric *models.Metrics

	if metricType == models.Counter {
		value, err := manager.metricService.CounterByNameAccumulative(metricID)
		if err != nil {
			return nil, err
		}
		modelMetric = &models.Metrics{
			MType: models.Counter,
			ID:    metricID,
			Delta: &value,
		}
	} else if metricType == models.Gauge {
		value, err := manager.metricService.GaugeByName(metricID)
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

func (manager *Manager) MetricList(metricType string) map[string]models.Metrics {

	metricsCounter := manager.metricService.CounterAccumulative()
	metricsGauge := manager.metricService.Gauge()

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

func (manager *Manager) MetricSnapshotLoad() error {
	return manager.metricSnapshot.Load()
}
func (manager *Manager) MetricSnapshotSave() error {
	return manager.metricSnapshot.Save()
}
