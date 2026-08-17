package server

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Env struct {
	cfg envCfg
}

type envCfg struct {
	Addr                  string `env:"ADDRESS"`
	MetricStoreInterval   *uint  `env:"STORE_INTERVAL"`
	MetricFileStoragePath string `env:"FILE_STORAGE_PATH"`
	MetricRestore         *bool  `env:"RESTORE"`
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

func (env *Env) Address() string {
	return env.cfg.Addr
}
func (env *Env) StoreInterval() *time.Duration {

	if env.cfg.MetricStoreInterval == nil {
		return nil
	}
	return new(time.Duration(*env.cfg.MetricStoreInterval) * time.Second)
}
func (env *Env) FileStoragePath() string {
	return env.cfg.MetricFileStoragePath
}
func (env *Env) Restore() *bool {
	return env.cfg.MetricRestore
}
