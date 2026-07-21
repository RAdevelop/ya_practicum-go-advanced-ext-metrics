package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate http.Handler
}

var metricStorage = service.NewMemStorage()

func New() *Handlers {

	metric := NewMetric(metricStorage)

	return &Handlers{
		MetricUpdate: MiddlewarePipeLine(http.HandlerFunc(metric.Update), MiddlewareValidator, MiddlewareContentTypeTextPlain, MiddlewareIsPostRequest),
	}
}
