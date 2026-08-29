package database

import (
	"context"
	"fmt"
	"strings"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/perrors"
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

var availableMetricTypes = map[string]bool{
	models.Gauge:   true,
	models.Counter: true,
}

func (s *Storage) UpdateBatch(ctx context.Context, metrics []models.Metrics) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(metrics) == 0 {
		return fmt.Errorf("%w", perrors.ErrMetricListEmpty)
	}

	var params = make([]any, 0, len(metrics)*4)
	var sqlValues = make([]string, 0, len(metrics))
	pHolderIndex := 1
	for _, metric := range metrics {
		params = append(params, metric.ID, metric.MType, metric.Delta, metric.Value)
		sqlValues = append(sqlValues, fmt.Sprintf("($%d, $%d, $%d, $%d)", pHolderIndex, pHolderIndex+1, pHolderIndex+2, pHolderIndex+3))
		pHolderIndex += 4
	}

	var queryBuilder strings.Builder
	queryBuilder.Grow(1024)
	queryBuilder.WriteString(`INSERT INTO metric (metric_id, m_type, delta, value) VALUES `)
	queryBuilder.WriteString(strings.Join(sqlValues, ","))
	queryBuilder.WriteString(`
			ON CONFLICT (metric_id, m_type)
			DO UPDATE SET
				-- Gauge: заменяем, если передано новое значение, иначе оставляем как было
				value = COALESCE(EXCLUDED.value, metric.value),
				-- Counter: суммируем, если передано новое значение, иначе оставляем как было
				delta = COALESCE(metric.delta + EXCLUDED.delta, metric.delta, EXCLUDED.delta),
				updated_at = CURRENT_TIMESTAMP
	`)

	_, err := s.DB.Executor(ctx).Exec(ctx, queryBuilder.String(), params...)
	if err != nil {
		return fmt.Errorf("%w, %w", perrors.ErrMetricUpdate, err)
	}
	return nil
}

func (s *Storage) Metric(ctx context.Context, metric *models.Metrics) (*models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if metric == nil {
		return nil, fmt.Errorf("%w", perrors.ErrMetricIsNil)
	}

	if !isTypeAvailable(metric.MType) {
		return nil, fmt.Errorf("%w", perrors.ErrMetricUnknownType)
	}

	if metric.ID == "" {
		return nil, fmt.Errorf("%w", perrors.ErrMetricEmptyID)
	}

	sql := `SELECT metric_id, m_type, delta, "value"  FROM metric WHERE metric_id = $1 AND m_type = $2`
	row := s.DB.Executor(ctx).QueryRow(ctx, sql, metric.ID, metric.MType)

	err := row.Scan(&metric.ID, &metric.MType, &metric.Delta, &metric.Value)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", perrors.ErrMetricNotFound, err)
	}

	return metric, nil
}

func (s *Storage) MetricList(ctx context.Context, metricType string) ([]models.Metrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if !isTypeAvailable(metricType) {
		return nil, fmt.Errorf("%w, metricType: %s", perrors.ErrMetricUnknownType, metricType)
	}

	sql := `SELECT metric_id, m_type, delta, "value", '' AS hash  FROM metric WHERE m_type = $1`

	rows, err := s.DB.Executor(ctx).Query(ctx, sql, metricType)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", perrors.ErrMetricNotFound, err)
	}

	metrics, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Metrics])

	if err != nil {
		return nil, fmt.Errorf("%w, %w", perrors.ErrMetricNotFound, err)
	}

	if len(metrics) == 0 {
		metrics = nil
		return nil, fmt.Errorf("%w", perrors.ErrMetricNotFound)
	}

	return metrics, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	return s.DB.Ping(ctx)
}

func isTypeAvailable(mType string) bool {
	if isAvailable, ok := availableMetricTypes[mType]; ok {
		return isAvailable
	}

	return false
}
