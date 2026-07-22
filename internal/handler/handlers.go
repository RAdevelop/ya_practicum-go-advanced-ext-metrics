package handler

import (
	"net/http"

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
		MetricUpdate: MiddlewarePipeLine(http.HandlerFunc(metric.Update), MiddlewareValidator, MiddlewareIsPostRequest),
		MetricGet:    MiddlewarePipeLine(http.HandlerFunc(metric.Get), MiddlewareValidator),
		MetricList:   http.HandlerFunc(metric.List),
	}
}
