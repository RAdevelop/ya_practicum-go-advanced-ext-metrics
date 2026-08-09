package main

import (
	"flag"
	"log"
	"net/http"

	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

func main() {

	var metricStorage = memory.NewStorage()
	var metricService = service.NewMetricService(metricStorage)

	var serverConfig configServer.ConfigProvider

	configServerEnv, err := configServer.NewEnv()
	if err != nil {
		log.Fatal(err)
		return
	}
	serverConfig = configServer.New(configServerEnv)

	srvAddress := flag.String("a", "localhost:8080", `Server address pattern: "host:port"`)
	flag.Parse()

	serverAddress := *srvAddress

	if serverConfig.Address() != "" {
		serverAddress = serverConfig.Address()
	}

	h := handler.New(metricService)
	r := router.New(h)

	err = http.ListenAndServe(serverAddress, r)
	if err != nil {
		log.Fatal(err)
	}
}
