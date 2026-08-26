package database

import (
	"context"
	"fmt"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate mockery
type Database interface {
	Close()
	Ping() error
}

type DB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

func New(ctx context.Context, cfg configDB.ConfigProvider) (*DB, error) {

	config, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.MaxConns() > 0 {
		config.MaxConns = cfg.MaxConns()
	}
	if cfg.MinConns() > 0 {
		config.MinConns = cfg.MinConns()
	}
	if cfg.MaxConnLifetime() > 0 {
		config.MaxConnLifetime = cfg.MaxConnLifetime()
	}
	if cfg.MaxConnIdleTime() > 0 {
		config.MaxConnIdleTime = cfg.MaxConnIdleTime()
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &DB{
		Pool: pool,
		ctx:  ctx,
	}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Ping - проверка доступности БД
func (db *DB) Ping() error {
	return db.Pool.Ping(db.ctx)
}
