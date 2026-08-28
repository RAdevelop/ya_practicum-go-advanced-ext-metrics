package database

import (
	"context"
	"errors"
	"os"
	"testing"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	"github.com/caarlos0/env/v11"
	"github.com/stretchr/testify/assert"
)

var envOpts = &env.Options{
	Environment: map[string]string{
		"DATABASE_DSN": os.Getenv("DB_DSN_TEST"),
	},
}

var errForTransactionRollback = errors.New("return for transaction rollback")

func setUpStorage(t *testing.T, ctx context.Context) *Storage {

	envDB, err := configDB.NewEnvWithOptions(envOpts)
	assert.NoError(t, err)

	//TODO del
	//t.Log("envDB.DSN()", envDB.DSN())

	db, err := NewDB(ctx, envDB)
	assert.NoError(t, err)

	t.Cleanup(func() {
		defer db.Close()
	})

	return NewStorage(db)
}

func TestStorage_GaugeUpdate(t *testing.T) {

	ctx := context.Background()
	storage := setUpStorage(t, ctx)

	mId := "metricNameRA"
	mValue := 0.0

	err := storage.DB.RunInTransaction(ctx, func(ctx context.Context) error {
		err := storage.GaugeUpdate(ctx, mId, mValue)
		assert.NoError(t, err)
		return errForTransactionRollback
	})

	assert.Error(t, err)
}
