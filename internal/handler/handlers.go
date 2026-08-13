package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate http.Handler
	MetricGet    http.Handler
	MetricList   http.Handler
}

func New(metricService *service.MetricService, logger logger.Logger) *Handlers {

	metric := NewMetric(metricService, logger)

	var metricUpdate = middleware.PipeLine(logger, http.HandlerFunc(metric.Update), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricGet = middleware.PipeLine(logger, http.HandlerFunc(metric.Get), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricList = middleware.PipeLine(logger, http.HandlerFunc(metric.List), middleware.Decompression, middleware.Compression, middleware.WithLogging)
	return &Handlers{
		MetricUpdate: metricUpdate,
		MetricGet:    metricGet,
		MetricList:   metricList,
	}
}
