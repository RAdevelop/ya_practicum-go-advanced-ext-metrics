package agent

import (
	"github.com/caarlos0/env/v11"
)

type Env struct {
	cfg envCfg
}
type envCfg struct {
	Addr           string `env:"ADDRESS"`
	IntervalReport uint   `env:"REPORT_INTERVAL"`
	IntervalPoll   uint   `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
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

func (env *Env) ReportInterval() uint {
	return env.cfg.IntervalReport
}

func (env *Env) PollInterval() uint {
	return env.cfg.IntervalPoll
}

func (env *Env) SignKey() string {
	return env.cfg.Key
}
