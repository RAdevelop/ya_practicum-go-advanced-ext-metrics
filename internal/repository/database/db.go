package database

import (
	"context"
	"fmt"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

//go:generate mockery
type Database interface {
	Close()
	Ping(context.Context) error
}

type ExecutorAble interface {
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

/*
DB - выполняем запросы к БД
ВАЖНО! Есть методы (contextWithTx, txFromContext), которые обновляют исходный контекст!
*/
type DB struct {
	pool *pgxpool.Pool
}

func NewDB(ctx context.Context, cfg configDB.ConfigProvider, logger logger.Logger) (*DB, error) {

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

	// Подключаем адаптер логгера
	if logger != nil {
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   newPgxLoggerAdapter(logger, tracelog.LogLevelError),
			LogLevel: tracelog.LogLevelError,
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return &DB{
		pool: pool,
	}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// Ping - проверка доступности БД
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// txKey — типизированный ключ, чтобы избежать коллизий в context.
type txKey struct{}

// contextWithTx кладёт транзакцию в контекст.
func (db *DB) contextWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// txFromContext достаёт транзакцию из контекста.
func (db *DB) txFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

func (db *DB) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Создаём контекст с транзакцией
	txCtx := db.contextWithTx(ctx, tx)

	// Передаём txCtx внутрь callback fn — именно его будут использовать storage.*
	if err = fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx) // Commit/Rollback используют исходный ctx для дедлайна или отмены
}

func (db *DB) Executor(ctx context.Context) ExecutorAble {
	// Пытаемся достать транзакцию из контекста
	if tx, ok := db.txFromContext(ctx); ok && tx != nil {
		return tx
	}
	// Иначе используем пул соединений
	return db.pool
}
