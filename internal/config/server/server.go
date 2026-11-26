// Package server - конфигурация сервера.
package server

import (
	"strings"
	"time"
)

// ConfigProvider - абстракция над работой настроек для сервера (они могут быть получены из: ENV переменных, INI файла, БД, и т.п.)
//
//go:generate mockery
type ConfigProvider interface {
	Address() string
	StoreInterval() *time.Duration
	FileStoragePath() string
	Restore() *bool
	SignKey() string
	AuditFile() string
	AuditURL() string
}

// Config - настройки для сервера
type Config struct {
	addr                  string
	metricStoreInterval   *time.Duration
	metricFileStoragePath string
	metricRestore         *bool
	signKey               string
	auditFile             string
	auditURL              string
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
	conf.SignKeySet(cfg.SignKey())
	conf.AuditFileSet(cfg.AuditFile())
	conf.AuditURLSet(cfg.AuditURL())

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

	fileStoragePath = strings.TrimSpace(fileStoragePath)
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

func (c *Config) SignKey() string {
	return c.signKey
}
func (c *Config) SignKeySet(signKey string) {
	signKey = strings.TrimSpace(signKey)
	if signKey != "" {
		c.signKey = signKey
	}
}

func (c *Config) AuditFile() string {
	return c.auditFile
}
func (c *Config) AuditFileSet(auditFile string) {
	auditFile = strings.TrimSpace(auditFile)
	if auditFile != "" {
		c.auditFile = auditFile
	}
}

func (c *Config) AuditURL() string {
	return c.auditURL
}
func (c *Config) AuditURLSet(auditURL string) {
	auditURL = strings.TrimSpace(auditURL)
	if auditURL != "" {
		c.auditURL = auditURL
	}
}
