package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

func main() {

	var metricStorage = memory.NewStorage()
	var metricService = service.NewMetricService(metricStorage)

	srvAddress := flag.String("a", "localhost:8080", `Server address pattern: "host:port"`)
	flag.Parse()

	h := handler.New(metricService)
	r := router.New(h)

	err := http.ListenAndServe(*srvAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
