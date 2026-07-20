package main

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
)

func main() {
	h := handler.New()

	mux := http.NewServeMux()

	handlerCounterUpdate := handler.MiddlewarePipeLine(http.HandlerFunc(h.Metric.Update), handler.MiddlewareValidator, handler.MiddlewareContentTypeTextPlain, handler.MiddlewareIsPostRequest)
	mux.Handle("/update/{metric_type}/{metric_name}/{metric_value}", handlerCounterUpdate)
	mux.Handle("/update/{metric_type}/{metric_name}", handlerCounterUpdate)
	mux.Handle("/update/{metric_type}", handlerCounterUpdate)
	mux.Handle("/update", handlerCounterUpdate)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
