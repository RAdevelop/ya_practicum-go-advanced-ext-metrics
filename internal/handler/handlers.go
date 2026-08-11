package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate http.Handler
	MetricGet    http.Handler
	MetricList   http.Handler
}

func New(metricService *service.MetricService) *Handlers {

	metric := NewMetric(metricService)

	return &Handlers{
		MetricUpdate: middleware.WithLogging(MiddlewarePipeLine(http.HandlerFunc(metric.Update), MiddlewareIsPostRequest)),
		MetricGet:    middleware.WithLogging(http.HandlerFunc(metric.Get)),
		MetricList:   middleware.WithLogging(http.HandlerFunc(metric.List)),
	}
}
