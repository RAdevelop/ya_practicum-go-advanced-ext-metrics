package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate      http.Handler
	MetricUpdateBatch http.Handler
	MetricGet         http.Handler
	MetricList        http.Handler
	MetricStoragePing http.Handler
}

func New(metricManager service.MetricManagementAble, serverContext *server.Context) *Handlers {

	metric := NewMetric(metricManager, serverContext)

	var metricUpdate = middleware.PipeLine(serverContext, http.HandlerFunc(metric.Update), middleware.SingCheck, middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricUpdateBatch = middleware.PipeLine(serverContext, http.HandlerFunc(metric.UpdateBatch), middleware.SingCheck, middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricGet = middleware.PipeLine(serverContext, http.HandlerFunc(metric.Get), middleware.SingCheck, middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricList = middleware.PipeLine(serverContext, http.HandlerFunc(metric.List), middleware.SingCheck, middleware.Decompression, middleware.Compression, middleware.WithLogging)
	var metricStoragePing = middleware.PipeLine(serverContext, http.HandlerFunc(metric.StoragePing), middleware.SingCheck, middleware.Decompression, middleware.Compression, middleware.WithLogging)

	return &Handlers{
		MetricUpdate:      metricUpdate,
		MetricUpdateBatch: metricUpdateBatch,
		MetricGet:         metricGet,
		MetricList:        metricList,
		MetricStoragePing: metricStoragePing,
	}
}
