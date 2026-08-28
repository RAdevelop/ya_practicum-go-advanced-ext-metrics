package database

import (
	"context"
	"errors"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/jackc/pgx/v5"
)

type Storage struct {
	DB *DB
}

func NewStorage(db *DB) *Storage {
	return &Storage{
		DB: db,
	}
}

func (s *Storage) GaugeUpdate(ctx context.Context, name string, value float64) error {
	return s.upsertMetric(ctx, models.Gauge, name, value)
}

func (s *Storage) GaugeByName(ctx context.Context, name string) (*models.Metrics, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.selectMetricRow(ctx, models.Gauge, name)
}

func (s *Storage) Gauge(ctx context.Context) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.selectMetricsRows(ctx, models.Gauge)
}

func (s *Storage) CounterAdd(ctx context.Context, name string, value int64) error {
	return s.upsertMetric(ctx, models.Counter, name, value)

}
func (s *Storage) CounterAccumulative(ctx context.Context) ([]models.Metrics, error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return s.selectMetricsRows(ctx, models.Counter)
}

func (s *Storage) CounterAccumulativeByName(ctx context.Context, name string) (*models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return s.selectMetricRow(ctx, models.Counter, name)
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}

// upsertMetric - добавляем метрику, или обновляем ее значение
func (s *Storage) upsertMetric(ctx context.Context, mType string, mID string, mValue any) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	var delta *int64
	var value *float64

	switch v := mValue.(type) {
	case float64:
		value = &v
	case int64:
		delta = &v
	}

	_, err := s.DB.Executor(ctx).Exec(ctx, `
		INSERT INTO metric (metric_id, m_type, delta, value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (metric_id, m_type) 
		DO UPDATE SET 
			-- Gauge: заменяем, если передано новое значение, иначе оставляем как было
			value = COALESCE(EXCLUDED.value, metric.value),
			-- Counter: суммируем, если передано новое значение, иначе оставляем как было
			delta = COALESCE(metric.delta + EXCLUDED.delta, metric.delta, EXCLUDED.delta),
			updated_at = CURRENT_TIMESTAMP`,
		mID, mType, delta, value,
	)
	return err
}

// selectMetricRow - ищем значения для метрики
func (s *Storage) selectMetricRow(ctx context.Context, mType string, mID string) (*models.Metrics, error) {

	sql := `SELECT metric_id, m_type, delta, "value"  FROM metric WHERE metric_id = $1 AND m_type = $2`

	row := s.DB.Executor(ctx).QueryRow(ctx, sql, mID, mType)

	var modelsMetric models.Metrics
	err := row.Scan(&modelsMetric.ID, &modelsMetric.MType, &modelsMetric.Delta, &modelsMetric.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("metric value not found for name: %s, %w", mID, err)
		}
		return nil, fmt.Errorf("scan error: %w", err)
	}

	return &modelsMetric, nil
}

// selectMetricsRows - ищем значения для метрик
func (s *Storage) selectMetricsRows(ctx context.Context, mType string) ([]models.Metrics, error) {

	sql := `SELECT metric_id, m_type, delta, "value" FROM metric WHERE m_type = $1`

	rows, err := s.DB.Executor(ctx).Query(ctx, sql, mType)
	if err != nil {
		return nil, err
	}

	metrics := make([]models.Metrics, 0, 30) // стоит написать COUNT() метод
	for rows.Next() {
		var modelMetric models.Metrics
		err = rows.Scan(&modelMetric.ID, &modelMetric.MType, &modelMetric.Delta, &modelMetric.Value)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, modelMetric)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(metrics) == 0 {
		metrics = nil
	}
	return metrics, nil
}
