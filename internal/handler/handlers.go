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

	return &Handlers{
		MetricUpdate: middleware.WithLogging(logger, http.HandlerFunc(metric.Update)),
		MetricGet:    middleware.WithLogging(logger, http.HandlerFunc(metric.Get)),
		MetricList:   middleware.WithLogging(logger, http.HandlerFunc(metric.List)),
	}
}
