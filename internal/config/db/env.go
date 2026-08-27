package db

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Env struct {
	cfg envCfg
}

type envCfg struct {
	DSN string `env:"DATABASE_DSN"`
	// Сколько всего подключений можно открыть к БД
	MaxConns int32 `env:"MAX_CONNS" envDefault:"25"`
	// Сколько соединений держать всегда открытыми
	MinConns int32 `env:"MIN_CONNS" envDefault:"5"`
	// Через сколько времени закрывать соединение, даже если оно используется
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME" envDefault:"1h"`
	// Через сколько времени закрывать соединение, если оно не используется
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME" envDefault:"4m"`
}

func NewEnv() (*Env, error) {
	return NewEnvWithOptions(nil)
}

// NewEnvWithOptions - Конструктор с опциями
func NewEnvWithOptions(opts *env.Options) (*Env, error) {
	var cfg envCfg
	var err error

	if opts != nil {
		err = env.ParseWithOptions(&cfg, *opts)
	} else {
		err = env.Parse(&cfg)
	}

	if err != nil {
		return nil, err
	}

	return &Env{
		cfg: cfg,
	}, nil
}

func (env *Env) DSN() string {
	return env.cfg.DSN
}

func (env *Env) MaxConns() int32 {
	return env.cfg.MaxConns
}
func (env *Env) MinConns() int32 {
	return env.cfg.MinConns
}

func (env *Env) MaxConnLifetime() time.Duration {
	return env.cfg.MaxConnLifetime
}
func (env *Env) MaxConnIdleTime() time.Duration {
	return env.cfg.MaxConnIdleTime
}
