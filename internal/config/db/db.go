package db

import "time"

// ConfigProvider - абстракция над работой настроек для DB (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
//
//go:generate mockery
type ConfigProvider interface {
	DSN() string
	MaxConns() int32
	MinConns() int32
	MaxConnLifetime() time.Duration
	MaxConnIdleTime() time.Duration
}

type Config struct {
	ConfigProvider  ConfigProvider
	dsn             string
	maxConns        int32
	minConns        int32
	maxConnLifetime time.Duration
	maxConnIdleTime time.Duration
}

func New(cfg ConfigProvider) *Config {
	conf := &Config{
		ConfigProvider: cfg,
	}
	conf.DSNSet(cfg.DSN())
	conf.maxConns = cfg.MaxConns()
	conf.minConns = cfg.MinConns()
	conf.maxConnLifetime = cfg.MaxConnLifetime()
	conf.maxConnIdleTime = cfg.MaxConnIdleTime()
	return conf
}

func (c *Config) DSN() string {
	return c.dsn
}

func (c *Config) DSNSet(dsn string) {
	c.dsn = dsn
}

// MaxConns - Сколько всего подключений можно открыть к БД
func (c *Config) MaxConns() int32 {
	return c.maxConns
}

// MinConns - Сколько соединений держать всегда открытыми
func (c *Config) MinConns() int32 {
	return c.minConns
}

// MaxConnLifetime - Через сколько времени закрывать соединение, даже если оно используется
func (c *Config) MaxConnLifetime() time.Duration {
	return c.maxConnLifetime
}

// MaxConnIdleTime - Через сколько времени закрывать соединение, если оно не используется
func (c *Config) MaxConnIdleTime() time.Duration {
	return c.maxConnIdleTime
}
