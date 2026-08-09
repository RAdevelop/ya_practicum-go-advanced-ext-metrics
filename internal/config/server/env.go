package server

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Env struct {
	Addr string `env:"ADDRESS"`
}

func NewEnv() *Env {

	var cfg Env
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return &cfg
}

func (env *Env) Address() string {
	return env.Addr
}
