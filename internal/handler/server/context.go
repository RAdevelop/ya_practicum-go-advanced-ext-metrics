package server

import (
	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

type Context struct {
	Logger logger.Logger
	Config configServer.ConfigProvider
}
