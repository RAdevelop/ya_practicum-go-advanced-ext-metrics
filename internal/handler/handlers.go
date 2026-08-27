package handler

import (
	"net/http"

	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate      http.Handler
	MetricGet         http.Handler
	MetricList        http.Handler
	MetricStoragePing http.Handler
}

func New(metricManager service.MetricManagementAble, logger logger.Logger, config configServer.ConfigProvider) *Handlers {

	metric := NewMetric(metricManager, logger, config)

	var metricUpdate = middleware.PipeLine(logger, http.HandlerFunc(metric.Update), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricGet = middleware.PipeLine(logger, http.HandlerFunc(metric.Get), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricList = middleware.PipeLine(logger, http.HandlerFunc(metric.List), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricStoragePing = middleware.PipeLine(logger, http.HandlerFunc(metric.StoragePing), middleware.Decompression, middleware.Compression, middleware.WithLogging)

	return &Handlers{
		MetricUpdate:      metricUpdate,
		MetricGet:         metricGet,
		MetricList:        metricList,
		MetricStoragePing: metricStoragePing,
	}
}
