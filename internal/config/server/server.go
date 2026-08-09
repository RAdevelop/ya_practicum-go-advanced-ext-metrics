package server

// ConfigProvider - абстракция над работой настроек для сервера (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
type ConfigProvider interface {
	Address() string
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
