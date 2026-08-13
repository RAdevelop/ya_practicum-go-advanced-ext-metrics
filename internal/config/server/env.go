package server

import (
	"github.com/caarlos0/env/v11"
)

type Env struct {
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
	var cfg Env
	var err error

	if opts != nil {
		err = env.ParseWithOptions(&cfg, *opts)
	} else {
		err = env.Parse(&cfg)
	}

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (env *Env) Address() string {
	return env.Addr
}
func (env *Env) StoreInterval() *uint {
	return env.MetricStoreInterval
}
func (env *Env) FileStoragePath() string {
	return env.MetricFileStoragePath
}
func (env *Env) Restore() *bool {
	return env.MetricRestore
}
