package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate http.Handler
}

var metricStorage = memory.NewMemStorage()
var metricService = service.NewMetricService(metricStorage)

func New() *Handlers {

	metric := NewMetric(metricService)

	return &Handlers{
		MetricUpdate: MiddlewarePipeLine(http.HandlerFunc(metric.Update), MiddlewareValidator, MiddlewareContentTypeTextPlain, MiddlewareIsPostRequest),
	}
}
