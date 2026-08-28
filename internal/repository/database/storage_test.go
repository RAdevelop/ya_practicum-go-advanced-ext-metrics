package database

import (
	"context"
	"errors"
	"os"
	"testing"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
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
	db, err := NewDB(ctx, envDB)
	assert.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
	})

	return NewStorage(db), ctx
}

func TestStorage_GaugeUpdateAndGaugeByName(t *testing.T) {

	type given struct {
		mId    string
		mValue float64
	}
	type want struct {
		hasError bool
		mValue   *models.Metrics
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "success",
			given: given{
				mId:    "gauge",
				mValue: 100.11,
			},
			want: want{
				hasError: false,
				mValue: &models.Metrics{
					ID:    "gauge",
					MType: models.Gauge,
					Value: new(100.11),
				},
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {
				err := storage.GaugeUpdate(ctx, tt.given.mId, tt.given.mValue)
				if tt.want.hasError {
					assert.Errorf(t, err, "gauge want: %v, given: %v", tt.want, tt.given)
				} else {
					assert.NoErrorf(t, err, "gauge want: %v, given: %v", tt.want, tt.given)
				}
				gaugeValue, err := storage.GaugeByName(ctx, tt.given.mId)
				assert.NoError(t, err)
				assert.Equalf(t, tt.want.mValue, gaugeValue, "gauge want: %v, given: %v", tt.want, tt.given)

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}
func TestStorage_Gauge(t *testing.T) {

	type given struct {
		gauge map[string]float64
	}
	type want struct {
		hasError bool
		gauge    []models.Metrics
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "success map not empty",
			given: given{
				gauge: map[string]float64{
					"gauge1": 101.11,
					"gauge2": 102.11,
				},
			},
			want: want{
				hasError: false,
				gauge: []models.Metrics{
					{
						ID:    "gauge1",
						MType: models.Gauge,
						Value: new(101.11),
					},
					{
						ID:    "gauge2",
						MType: models.Gauge,
						Value: new(102.11),
					},
				},
			},
		},
		{
			name: "success map is empty",
			given: given{
				gauge: map[string]float64{},
			},
			want: want{
				hasError: false,
				gauge:    nil,
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {
				for mId, mValue := range tt.given.gauge {
					err := storage.GaugeUpdate(ctx, mId, mValue)
					if tt.want.hasError {
						assert.Errorf(t, err, "gauge want: %v, given: %v", tt.want, tt.given)
					} else {
						assert.NoErrorf(t, err, "gauge want: %v, given: %v", tt.want, tt.given)
					}
				}

				gaugeValue, err := storage.Gauge(ctx)

				t.Logf("--------------------------------------------------------------gaugeValue: %v", gaugeValue)
				assert.NoError(t, err)
				assert.Equalf(t, tt.want.gauge, gaugeValue, "gauge want: %v, given: %v", tt.want, tt.given)

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}

func TestStorage_CounterAddAndCounterAccumulativeByName(t *testing.T) {

	type given struct {
		mId    string
		mValue int64
	}
	type want struct {
		hasError bool
		mValue   *models.Metrics
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "success",
			given: given{
				mId:    "counter",
				mValue: 100,
			},
			want: want{
				hasError: false,
				mValue: &models.Metrics{
					ID:    "counter",
					MType: models.Counter,
					Delta: new(int64(100)),
				},
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {

				err := storage.CounterAdd(ctx, tt.given.mId, tt.given.mValue)
				if tt.want.hasError {
					assert.Errorf(t, err, "counter want: %v, given: %v", tt.want, tt.given)
				} else {
					assert.NoErrorf(t, err, "counter want: %v, given: %v", tt.want, tt.given)
				}

				counterValue, err := storage.CounterAccumulativeByName(ctx, tt.given.mId)

				assert.NoError(t, err)
				assert.Equalf(t, tt.want.mValue, counterValue, "counter want: %v, given: %v", tt.want, tt.given)

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}

func TestStorage_CounterAccumulative(t *testing.T) {

	type given struct {
		counter map[string]int64
	}
	type want struct {
		hasError bool
		counter  []models.Metrics
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "success map not empty",
			given: given{
				counter: map[string]int64{
					"counter1": 100,
					"counter2": 101,
				},
			},
			want: want{
				hasError: false,
				counter: []models.Metrics{
					{
						ID:    "counter1",
						MType: models.Counter,
						Delta: new(int64(100)),
					},
					{
						ID:    "counter2",
						MType: models.Counter,
						Delta: new(int64(101)),
					},
				},
			},
		},
		{
			name: "success map is empty",
			given: given{
				counter: map[string]int64{},
			},
			want: want{
				hasError: false,
				counter:  []models.Metrics{},
			},
		},
	}

	storage, ctx := setUpStorage(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {

				for mId, mValue := range tt.given.counter {
					err := storage.CounterAdd(ctx, mId, mValue)
					if tt.want.hasError {
						assert.Errorf(t, err, "counter want: %v, given: %v", tt.want, tt.given)
					} else {
						assert.NoErrorf(t, err, "counter want: %v, given: %v", tt.want, tt.given)
					}
				}

				counterValue, err := storage.CounterAccumulative(ctx)

				t.Logf("---------counterValue: %v", counterValue)

				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.want.counter, counterValue, "counter want: %v, given: %v", tt.want, tt.given)

				return errForTransactionRollback
			})

			assert.Error(t, err)
		})
	}
}
