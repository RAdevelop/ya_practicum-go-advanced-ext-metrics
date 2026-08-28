package database

import (
	"context"
	"fmt"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
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
	return s.upsert(ctx, models.Gauge, name, value)
}

func (s *Storage) GaugeByName(ctx context.Context, name string) (float64, error) {
	return 0.0, fmt.Errorf("GaugeByName not implemented")
}

func (s *Storage) Gauge(ctx context.Context) map[string]float64 {
	return make(map[string]float64) //TODO
}

func (s *Storage) CounterAdd(ctx context.Context, name string, value int64) error {
	return fmt.Errorf("CounterAdd not implemented")

}
func (s *Storage) CounterAccumulative(ctx context.Context) map[string]int64 {
	return make(map[string]int64) //TODO
}

func (s *Storage) CounterAccumulativeByName(ctx context.Context, name string) (int64, error) {
	return 0, fmt.Errorf("CounterAccumulativeByName not implemented")
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}

// upsert - добавляем метрику, или обновляем ее значение
func (s *Storage) upsert(ctx context.Context, mType string, mID string, mValue any) error {

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
	default:
		return fmt.Errorf("unsupported type: %T", mValue)
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
