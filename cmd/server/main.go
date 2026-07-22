package main

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

var metricStorage = memory.NewStorage()
var metricService = service.NewMetricService(metricStorage)

func main() {

	h := handler.New(metricService)
	r := router.New(h)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		panic(err)
	}
}
