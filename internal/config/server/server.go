package server

// ConfigProvider - абстракция над работой настроек для сервера (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
type ConfigProvider interface {
	Address() string
	StoreInterval() *uint
	FileStoragePath() string
	Restore() *bool
}

// Config - настройки для сервера
type Config struct {
	cfgProvider ConfigProvider
}

func New(cfg ConfigProvider) *Config {
	return &Config{
		cfgProvider: cfg,
	}
}

// Address - отвечает за адрес эндпоинта HTTP-сервера.
func (c *Config) Address() string {
	return c.cfgProvider.Address()
}

func (c *Config) StoreInterval() *uint {
	return c.cfgProvider.StoreInterval()
}

func (c *Config) FileStoragePath() string {
	return c.cfgProvider.FileStoragePath()
}

func (c *Config) Restore() *bool {
	return c.cfgProvider.Restore()
}
