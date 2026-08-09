package agent

import (
	"github.com/caarlos0/env/v11"
)

type Env struct {
	Addr           string `env:"ADDRESS"`
	IntervalReport uint   `env:"REPORT_INTERVAL"`
	IntervalPoll   uint   `env:"POLL_INTERVAL"`
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

func (env *Env) ReportInterval() uint {
	return env.IntervalReport
}

func (env *Env) PollInterval() uint {
	return env.IntervalPoll
}
