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

	middlewares := []middleware.Middleware{
		middleware.SignCheck,
		middleware.Decompression,
		middleware.Compression,
		middleware.SignAdd,
		middleware.WithLogging,
	}

	var metricUpdate = middleware.PipeLine(serverContext, http.HandlerFunc(metric.Update), middlewares...)
	var metricUpdateBatch = middleware.PipeLine(serverContext, http.HandlerFunc(metric.UpdateBatch), middlewares...)
	var metricGet = middleware.PipeLine(serverContext, http.HandlerFunc(metric.Get), middlewares...)
	var metricList = middleware.PipeLine(serverContext, http.HandlerFunc(metric.List), middlewares...)
	var metricStoragePing = middleware.PipeLine(serverContext, http.HandlerFunc(metric.StoragePing), middlewares...)

	return &Handlers{
		MetricUpdate:      metricUpdate,
		MetricUpdateBatch: metricUpdateBatch,
		MetricGet:         metricGet,
		MetricList:        metricList,
		MetricStoragePing: metricStoragePing,
	}
}
