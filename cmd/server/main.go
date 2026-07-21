package main

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
)

func main() {
	h := handler.New()

	mux := http.NewServeMux()

	mux.Handle("/update/{metric_type}/{metric_name}/{metric_value}", h.MetricUpdate)
	mux.Handle("/update/{metric_type}/{metric_name}", h.MetricUpdate)
	mux.Handle("/update/{metric_type}", h.MetricUpdate)
	mux.Handle("/update", h.MetricUpdate)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
