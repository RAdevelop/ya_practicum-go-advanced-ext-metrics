package repository

import (
	"context"
	"errors"
	"os"
	"testing"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/perrors"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/database"
	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

var envOpts = &env.Options{
	Environment: map[string]string{
		"DATABASE_DSN": os.Getenv("DB_DSN_TEST"),
	},
}

var errForTransactionRollback = errors.New("return for transaction rollback")

func setUpStorage(t *testing.T) (*Storage, context.Context) {

	envDB, err := configDB.NewEnvWithOptions(envOpts)
	assert.NoError(t, err)

	ctx := context.Background()
	db, err := database.NewDB(ctx, envDB, nil)
	assert.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return NewDatabase(db), ctx
}

func TestStorage_UpdateBatch(t *testing.T) {

	type given struct {
		metrics []models.Metrics
	}
	type want struct {
		err error
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "ErrMetricListEmpty nil",
			given: given{
				metrics: nil,
			},
			want: want{
				err: perrors.ErrMetricListEmpty,
			},
		},
		{
			name: "ErrMetricListEmpty",
			given: given{
				metrics: []models.Metrics{},
			},
			want: want{
				err: perrors.ErrMetricListEmpty,
			},
		},
		{
			name: "No Error",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(123)),
					},
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(123)),
					},
				},
			},
			want: want{
				err: nil,
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {

				_, err := storage.UpdateBatch(ctx, tt.given.metrics)
				if tt.want.err != nil {
					assert.ErrorIs(t, err, tt.want.err)
				} else {
					assert.NoError(t, err)
				}

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}

func TestStorage_Metric(t *testing.T) {

	type given struct {
		metrics []models.Metrics
		metric  *models.Metrics
	}
	type want struct {
		err    error
		metric *models.Metrics
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "ErrMetricIsNil",
			given: given{
				metrics: nil,
				metric:  nil,
			},
			want: want{
				err:    perrors.ErrMetricIsNil,
				metric: nil,
			},
		},
		{
			name: "ErrMetricUnknownType",
			given: given{
				metrics: nil,
				metric: &models.Metrics{
					ID:    "Counter",
					MType: "UnknownType",
				},
			},
			want: want{
				err:    perrors.ErrMetricUnknownType,
				metric: nil,
			},
		},
		{
			name: "ErrMetricEmptyID",
			given: given{
				metrics: nil,
				metric: &models.Metrics{
					ID:    "",
					MType: models.Counter,
				},
			},
			want: want{
				err:    perrors.ErrMetricEmptyID,
				metric: nil,
			},
		},
		{
			name: "not found Counter",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(123)),
					},
				},
				metric: &models.Metrics{
					ID:    "NotFoundCounter",
					MType: models.Counter,
				},
			},
			want: want{
				err:    perrors.ErrMetricNotFound,
				metric: nil,
			},
		},
		{
			name: "not found Gauge",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(123)),
					},
				},
				metric: &models.Metrics{
					ID:    "NotFoundGauge",
					MType: models.Gauge,
				},
			},
			want: want{
				err:    perrors.ErrMetricNotFound,
				metric: nil,
			},
		},
		{
			name: "success get Counter",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(123)),
					},
				},
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
				},
			},
			want: want{
				err: nil,
				metric: &models.Metrics{
					ID:    "Counter",
					MType: models.Counter,
					Delta: new(int64(123)),
				},
			},
		},
		{
			name: "success get Gauge",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(123)),
					},
				},
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
				},
			},
			want: want{
				err: nil,
				metric: &models.Metrics{
					ID:    "Gauge",
					MType: models.Gauge,
					Value: new(float64(123)),
				},
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {

				_, _ = storage.UpdateBatch(ctx, tt.given.metrics)

				metric, err := storage.Metric(ctx, tt.given.metric)

				if tt.want.err != nil {
					assert.ErrorIs(t, err, tt.want.err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tt.want.metric, metric)

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}

func TestStorage_MetricList(t *testing.T) {
	type given struct {
		metrics    []models.Metrics
		metricType string
	}
	type want struct {
		err     error
		metrics []models.Metrics
	}
	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "ErrMetricUnknownType",
			given: given{
				metrics:    nil,
				metricType: "UnknownType",
			},
			want: want{
				err:     perrors.ErrMetricUnknownType,
				metrics: nil,
			},
		},
		{
			name: "ErrMetricNotFound Counter",
			given: given{
				metrics:    nil,
				metricType: models.Counter,
			},
			want: want{
				err:     perrors.ErrMetricNotFound,
				metrics: nil,
			},
		},
		{
			name: "ErrMetricNotFound Gauge",
			given: given{
				metrics:    nil,
				metricType: models.Gauge,
			},
			want: want{
				err:     perrors.ErrMetricNotFound,
				metrics: nil,
			},
		},
		{
			name: "found Gauge",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(123)),
					},
				},
				metricType: models.Gauge,
			},
			want: want{
				err: nil,
				metrics: []models.Metrics{
					{
						ID:    "Gauge",
						MType: models.Gauge,
						Value: new(float64(123)),
					},
				},
			},
		},
		{
			name: "found Counter",
			given: given{
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(123)),
					},
				},
				metricType: models.Counter,
			},
			want: want{
				err: nil,
				metrics: []models.Metrics{
					{
						ID:    "Counter",
						MType: models.Counter,
						Delta: new(int64(123)),
					},
				},
			},
		},
	}

	storage, ctx := setUpStorage(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {

				_, _ = storage.UpdateBatch(ctx, tt.given.metrics)

				metrics, err := storage.MetricList(ctx, tt.given.metricType)

				if tt.want.err != nil {
					assert.ErrorIs(t, err, tt.want.err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tt.want.metrics, metrics)

				return errForTransactionRollback
			})
			assert.Error(t, err)
		})
	}
}

func TestStorage_Ping(t *testing.T) {

	type given struct {
		envOpts *env.Options
	}
	type want struct {
		hasError bool
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "ErrPing",
			given: given{
				envOpts: &env.Options{},
			},
			want: want{
				hasError: true,
			},
		},
		{
			name: "NoErrorPing",
			given: given{
				envOpts: &env.Options{
					Environment: map[string]string{
						"DATABASE_DSN": os.Getenv("DB_DSN_TEST"),
					},
				},
			},
			want: want{
				hasError: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			envDB, err := configDB.NewEnvWithOptions(tt.given.envOpts)
			assert.NoError(t, err)

			ctx := context.Background()
			db, err := database.NewDB(ctx, envDB, nil)
			assert.NoError(t, err)

			t.Cleanup(func() {
				db.Close()
			})

			storage := NewDatabase(db)

			err = storage.Ping(ctx)
			if tt.want.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
