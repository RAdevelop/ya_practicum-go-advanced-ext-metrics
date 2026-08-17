package server

import "time"

// ConfigProvider - абстракция над работой настроек для сервера (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
type ConfigProvider interface {
	Address() string
	StoreInterval() *time.Duration
	FileStoragePath() string
	Restore() *bool
}

// Config - настройки для сервера
type Config struct {
	addr                  string
	metricStoreInterval   *time.Duration
	metricFileStoragePath string
	metricRestore         *bool
}

func New(cfg ConfigProvider) *Config {

	conf := &Config{}

	conf.AddressSet(cfg.Address())
	var interval *uint
	if cfg.StoreInterval() != nil {
		interval = new(uint((*cfg.StoreInterval()) / time.Second))
	}
	conf.StoreIntervalSet(interval)

	conf.FileStoragePathSet(cfg.FileStoragePath())
	conf.RestoreSet(cfg.Restore())

	return conf
}

// Address - отвечает за адрес эндпоинта HTTP-сервера.
func (c *Config) Address() string {
	return c.addr
}
func (c *Config) AddressSet(addr string) {
	if addr != "" {
		c.addr = addr
	}
}

func (c *Config) StoreInterval() *time.Duration {
	return c.metricStoreInterval
}
func (c *Config) StoreIntervalSet(storeInterval *uint) {

	if storeInterval != nil {
		c.metricStoreInterval = new(time.Duration(*storeInterval) * time.Second)
	}
}

func (c *Config) FileStoragePath() string {
	return c.metricFileStoragePath
}

func (c *Config) FileStoragePathSet(fileStoragePath string) {
	if fileStoragePath != "" {
		c.metricFileStoragePath = fileStoragePath
	}
}

func (c *Config) Restore() *bool {
	return c.metricRestore
}
func (c *Config) RestoreSet(metricRestore *bool) {
	if metricRestore != nil {
		c.metricRestore = metricRestore
	}
}
