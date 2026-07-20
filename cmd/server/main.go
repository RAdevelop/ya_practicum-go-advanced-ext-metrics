package main

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
)

func main() {
	h := handler.New()

	mux := http.NewServeMux()

	handlerMetricUpdate := handler.MiddlewarePipeLine(http.HandlerFunc(h.Metric.Update), handler.MiddlewareValidator, handler.MiddlewareContentTypeTextPlain, handler.MiddlewareIsPostRequest)
	mux.Handle("/update/{metric_type}/{metric_name}/{metric_value}", handlerMetricUpdate)
	mux.Handle("/update/{metric_type}/{metric_name}", handlerMetricUpdate)
	mux.Handle("/update/{metric_type}", handlerMetricUpdate)
	mux.Handle("/update", handlerMetricUpdate)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
