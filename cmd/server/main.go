package main

import (
	"flag"
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

var metricStorage = memory.NewStorage()
var metricService = service.NewMetricService(metricStorage)

func main() {

	srvAddress := &serverAddress{
		host: "localhost",
		port: 8080,
	}
	_ = flag.Value(srvAddress)
	flag.Var(srvAddress, "a", `Server address pattern: "host:port"`)
	flag.Parse()

	h := handler.New(metricService)
	r := router.New(h)

	err := http.ListenAndServe(srvAddress.String(), r)
	if err != nil {
		panic(err)
	}
}
