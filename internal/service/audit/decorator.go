package audit

import (
	"context"
	"time"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

// Decorator — обёртка над MetricManagementAble, публикующая события аудита после успешных сохранений метрик.
type Decorator struct {
	manager   service.MetricManagementAble
	publisher *Publisher
}

func NewDecorator(manager service.MetricManagementAble, publisher *Publisher) *Decorator {
	return &Decorator{
		manager:   manager,
		publisher: publisher,
	}
}

// MetricUpdateBatch — выполняет аудит после успешного обновления метрик.
func (d *Decorator) MetricUpdateBatch(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	result, err := d.manager.MetricUpdateBatch(ctx, metrics)
	if err != nil {
		return result, err
	}
	event := Event{
		TS:      time.Now().Unix(),
		Metrics: metricNames(metrics),
		IP:      IPFromContext(ctx),
	}

	d.publisher.Notify(ctx, event)

	return result, nil
}

func (d *Decorator) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	return d.manager.Metric(ctx, metric)
}

func (d *Decorator) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	return d.manager.MetricList(ctx, metricType)
}

func (d *Decorator) MetricSnapshotLoad(ctx context.Context) error {
	return d.manager.MetricSnapshotLoad(ctx)
}

func (d *Decorator) MetricSnapshotSave(ctx context.Context) error {
	return d.manager.MetricSnapshotSave(ctx)
}

func (d *Decorator) StoragePing(ctx context.Context) error {
	return d.manager.StoragePing(ctx)
}

// metricNames — вытаскивает имена метрик из списка.
func metricNames(metrics []models.Metrics) []string {
	names := make([]string, 0, len(metrics))
	for _, m := range metrics {
		names = append(names, m.ID)
	}
	return names
}
