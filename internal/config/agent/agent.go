package agent

// ConfigProvider - абстракция над работой настроек для агента (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
type ConfigProvider interface {
	Address() string
	ReportInterval() uint
	PollInterval() uint
}

type Config struct {
	cfgProvider ConfigProvider
}

func New(cfgProvider ConfigProvider) *Config {
	return &Config{cfgProvider: cfgProvider}
}

// ReportInterval - позволяет переопределять `reportInterval`.
func (c *Config) ReportInterval() uint {
	return c.cfgProvider.ReportInterval()
}

// PollInterval - позволяет переопределять `pollInterval`.
func (c *Config) PollInterval() uint {
	return c.cfgProvider.PollInterval()
}

// Address - отвечает за адрес эндпоинта HTTP-сервера.
func (c *Config) Address() string {
	return c.cfgProvider.Address()
}
