package agent

// ConfigProvider - абстракция над работой настроек для агента (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
type ConfigProvider interface {
	Address() string
	ReportInterval() uint
	PollInterval() uint
	SignKey() string
	RateLimit() uint
}

type Config struct {
	addr           string
	intervalReport uint
	intervalPoll   uint
	signKey        string
	rateLimit      uint
}

func New(cfgProvider ConfigProvider) *Config {

	cfg := &Config{}

	cfg.addr = cfgProvider.Address()
	cfg.intervalReport = cfgProvider.ReportInterval()
	cfg.intervalPoll = cfgProvider.PollInterval()
	cfg.signKey = cfgProvider.SignKey()

	return cfg
}

// ReportInterval - позволяет переопределять `reportInterval`.
func (c *Config) ReportInterval() uint {
	return c.intervalReport
}
func (c *Config) ReportIntervalSet(intervalReport uint) {
	c.intervalReport = intervalReport
}

// PollInterval - позволяет переопределять `pollInterval`.
func (c *Config) PollInterval() uint {
	return c.intervalPoll
}
func (c *Config) PollIntervalSet(intervalPoll uint) {
	c.intervalPoll = intervalPoll
}

// Address - отвечает за адрес эндпоинта HTTP-сервера.
func (c *Config) Address() string {
	return c.addr
}
func (c *Config) AddressSet(addr string) {
	c.addr = addr
}

// SignKey - отвечает за ключ для подписи запроса
func (c *Config) SignKey() string {
	return c.signKey
}
func (c *Config) SignKeySet(signKey string) {
	c.signKey = signKey
}
func (c *Config) RateLimit() uint {
	return c.rateLimit
}
func (c *Config) RateLimitSet(rateLimit uint) {
	c.rateLimit = rateLimit
}
