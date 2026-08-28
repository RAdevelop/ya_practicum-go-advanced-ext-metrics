package service

import (
	"context"
	"errors"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
)

/*
ErrUnknownMetricType - неизвестный тип метрики

ВОПРОС К РЕВЬЮЕРУ: в разных пакетах иногда приходится объявлять одинаковые по смыслу ошибку.
Вопрос - насколько в Go практикуется создавать своего рода общий пакет с объявлением ошибок, и уже его использовать в коде,
чтобы не плодить идентичные по смыслу ошибку?
*/
var ErrUnknownMetricType = errors.New("unknown metric type")

type MetricManagementAble interface {
	MetricUpdate(context.Context, *models.Metrics) error
	MetricValue(context.Context, string, string) (*models.Metrics, error)
	MetricList(context.Context, string) ([]models.Metrics, error)
	MetricSnapshotLoad(context.Context) error
	MetricSnapshotSave(context.Context) error
	StoragePing(context.Context) error
}

// Manager - предоставляет интерфейс (фасад) для работы с метриками
type Manager struct {
	storage        metric.Storage
	metricSnapshot snapshot.Able
}

func NewManager(storage metric.Storage, metricSnapshot snapshot.Able) *Manager {
	return &Manager{
		storage:        storage,
		metricSnapshot: metricSnapshot,
	}
}

func (manager *Manager) MetricUpdate(ctx context.Context, metric *models.Metrics) error {
	if metric.MType == models.Counter && metric.Delta != nil {
		return manager.storage.CounterAdd(ctx, metric.ID, *metric.Delta)
	} else if metric.MType == models.Gauge && metric.Value != nil {
		return manager.storage.GaugeUpdate(ctx, metric.ID, *metric.Value)
	}

	return fmt.Errorf("%w: %s", ErrUnknownMetricType, metric.MType)
}

func (manager *Manager) MetricValue(ctx context.Context, metricType string, metricID string) (*models.Metrics, error) {
	var modelMetric *models.Metrics
	var err error
	if metricType == models.Counter {
		modelMetric, err = manager.storage.CounterAccumulativeByName(ctx, metricID)
		if err != nil {
			return nil, err
		}
	} else if metricType == models.Gauge {
		modelMetric, err = manager.storage.GaugeByName(ctx, metricID)
		if err != nil {
			return nil, err
		}
	}

	return modelMetric, nil
}

func (manager *Manager) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	var metrics []models.Metrics

	if metricType == models.Counter {
		metricsCounter, err := manager.storage.CounterAccumulative(ctx)
		if err != nil {
			return nil, err
		}
		metrics = make([]models.Metrics, 0, len(metricsCounter))
		for _, value := range metricsCounter {
			metrics = append(metrics, value)
		}
	} else if metricType == models.Gauge {
		metricsGauge, err := manager.storage.Gauge(ctx)
		if err != nil {
			return nil, err
		}
		metrics = make([]models.Metrics, 0, len(metricsGauge))
		for _, value := range metricsGauge {
			metrics = append(metrics, value)
		}
	}

	return metrics, nil
}

func (manager *Manager) MetricSnapshotLoad(ctx context.Context) error {
	return manager.metricSnapshot.Load(ctx)
}
func (manager *Manager) MetricSnapshotSave(ctx context.Context) error {
	return manager.metricSnapshot.Save(ctx)
}

func (manager *Manager) StoragePing(ctx context.Context) error {
	return manager.storage.Ping(ctx)
}
