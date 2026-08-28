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

func (s *Storage) GaugeByName(ctx context.Context, name string) (float64, error) {

	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var value *float64
	row := s.selectMetricRow(ctx, models.Gauge, name, "value")

	err := row.Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("metric value not found for name: %s, %w", name, err)
		}
		return 0, fmt.Errorf("scan error: %w", err)
	}

	return *value, err
}

func (s *Storage) Gauge(ctx context.Context) map[string]float64 {

	result := make(map[string]float64)

	if err := ctx.Err(); err != nil {
		return result
	}

	rows, err := s.selectMetricsRows(ctx, models.Gauge, "value")
	if err != nil {
		return result
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return result
		}
		result[name] = value
	}

	//TODO надо менять интерфейс, чтобы он возвращал карту models.Metrics и error
	if err = rows.Err(); err != nil {
		return result
	}
	return result
}

func (s *Storage) CounterAdd(ctx context.Context, name string, value int64) error {
	return s.upsertMetric(ctx, models.Counter, name, value)

}
func (s *Storage) CounterAccumulative(ctx context.Context) map[string]int64 {
	result := make(map[string]int64)

	if err := ctx.Err(); err != nil {
		return result
	}

	rows, err := s.selectMetricsRows(ctx, models.Counter, "delta")
	if err != nil {
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return result
		}
		result[name] = value
	}

	//TODO надо менять интерфейс, чтобы он возвращал карту models.Metrics и error
	if err = rows.Err(); err != nil {
		return result
	}
	return result
}

func (s *Storage) CounterAccumulativeByName(ctx context.Context, name string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	var delta *int64
	row := s.selectMetricRow(ctx, models.Counter, name, "delta")

	err := row.Scan(&delta)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("metric value not found for name: %s, %w", name, err)
		}
		return 0, fmt.Errorf("scan error: %w", err)
	}

	return *delta, err
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
func (s *Storage) selectMetricRow(ctx context.Context, mType string, mID string, field string) pgx.Row {

	sql := `SELECT ` + field + ` FROM metric WHERE metric_id = $1 AND m_type = $2`

	return s.DB.Executor(ctx).QueryRow(ctx, sql, mID, mType)
}

// selectMetricsRows - ищем значения для метрик
func (s *Storage) selectMetricsRows(ctx context.Context, mType string, field string) (pgx.Rows, error) {

	sql := `SELECT metric_id, ` + field + ` FROM metric WHERE m_type = $1`

	return s.DB.Executor(ctx).Query(ctx, sql, mType)
}
