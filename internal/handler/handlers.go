package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/appcontext"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

type Handlers struct {
	MetricUpdate      http.Handler
	MetricUpdateBatch http.Handler
	MetricGet         http.Handler
	MetricList        http.Handler
	MetricStoragePing http.Handler
}

func New(metricManager service.MetricManagementAble, appContext *appcontext.AppContext) *Handlers {

	metric := NewMetric(metricManager, appContext)

	middlewaresWithSignCheck := []middleware.Middleware{
		middleware.SignCheck,
		middleware.Decompression,
		middleware.Compression,
		middleware.SignAdd,
		middleware.WithLogging,
	}

	middlewaresWithOutSignCheck := []middleware.Middleware{
		middleware.Decompression,
		middleware.Compression,
		middleware.SignAdd,
		middleware.WithLogging,
	}

	var metricUpdate = middleware.PipeLine(appContext, http.HandlerFunc(metric.Update), middlewaresWithOutSignCheck...)
	var metricUpdateBatch = middleware.PipeLine(appContext, http.HandlerFunc(metric.UpdateBatch), middlewaresWithSignCheck...)
	var metricGet = middleware.PipeLine(appContext, http.HandlerFunc(metric.Get), middlewaresWithOutSignCheck...)
	var metricList = middleware.PipeLine(appContext, http.HandlerFunc(metric.List), middlewaresWithOutSignCheck...)
	var metricStoragePing = middleware.PipeLine(appContext, http.HandlerFunc(metric.StoragePing), middlewaresWithOutSignCheck...)

	return &Handlers{
		MetricUpdate:      metricUpdate,
		MetricUpdateBatch: metricUpdateBatch,
		MetricGet:         metricGet,
		MetricList:        metricList,
		MetricStoragePing: metricStoragePing,
	}
}
