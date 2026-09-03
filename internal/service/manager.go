package service

import (
	"context"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/retryer"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
)

type MetricManagementAble interface {
	MetricUpdateBatch(context.Context, []models.Metrics) ([]models.Metrics, error)
	Metric(context.Context, *models.Metrics) (*models.Metrics, error)
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

func (manager *Manager) MetricUpdateBatch(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	return retryer.RetryLinear(ctx, func(ctx context.Context) ([]models.Metrics, error) {
		return manager.storage.UpdateBatch(ctx, metrics)
	}, 2, new(3))
}

func (manager *Manager) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	return retryer.RetryLinear(ctx, func(ctx context.Context) (*models.Metrics, error) {
		return manager.storage.Metric(ctx, metric)
	}, 2, new(3))
}

func (manager *Manager) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	return retryer.RetryLinear(ctx, func(ctx context.Context) ([]models.Metrics, error) {
		return manager.storage.MetricList(ctx, metricType)
	}, 2, new(3))
}

func (manager *Manager) MetricSnapshotLoad(ctx context.Context) error {
	return manager.metricSnapshot.Load(ctx)
}
func (manager *Manager) MetricSnapshotSave(ctx context.Context) error {
	return manager.metricSnapshot.Save(ctx)
}

func (manager *Manager) StoragePing(ctx context.Context) error {
	var err error
	_, err = retryer.RetryLinear(ctx, func(ctx context.Context) (struct{}, error) {

		err = manager.storage.Ping(ctx)
		return struct{}{}, err
	}, 2, new(3))
	return err
}
