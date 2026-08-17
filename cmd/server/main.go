package main

import (
	"flag"
	"net/http"
	"time"

	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
)

func main() {

	logMe := logger.New()

	configServerEnv, err := configServer.NewEnv()
	if err != nil {
		logMe.Error("error", "err", err)
		return
	}
	serverConfig := configServer.New(configServerEnv)

	srvAddress := flag.String("a", "localhost:8080", `Server address pattern: "host:port"`)
	srvStoreInterval := flag.Uint("i", 300, `интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)`)
	srvFileStoragePath := flag.String("f", "dump/metrics/iter9.json", `путь до файла, куда сохраняются текущие значения`)
	srvRestore := flag.Bool("r", true, `булево значение (true/false), определяющее, следует ли загружать ранее сохранённые значения из указанного файла при старте сервера.`)

	flag.Parse()

	if serverConfig.Address() == "" {
		serverConfig.AddressSet(*srvAddress)
	}

	if serverConfig.StoreInterval() == nil {
		serverConfig.StoreIntervalSet(srvStoreInterval)
	}
	if serverConfig.FileStoragePath() == "" {
		serverConfig.FileStoragePathSet(*srvFileStoragePath)
	}

	if serverConfig.Restore() == nil {
		serverConfig.RestoreSet(srvRestore)
	}

	var metricStorage = memory.NewStorage()
	var metricService = service.NewMetric(metricStorage)
	h := handler.New(metricService, logMe)
	r := router.New(h)

	metricInitializer, err := service.NewMetricInitializer(serverConfig.FileStoragePath(), metricService)
	if err != nil {
		logMe.Error("error", "err", err)
		return
	}

	if serverConfig.Restore() != nil && *serverConfig.Restore() {
		err = metricInitializer.Load()
		if err != nil {
			logMe.Error("metricSaver Load", "err", err)
		}
	}

	go saver(metricInitializer, logMe, serverConfig)

	err = http.ListenAndServe(serverConfig.Address(), r)
	if err != nil {
		logMe.Error("error", "err", err)
	}
}

func saver(metricInitializer *service.MetricInitializer, logger logger.Logger, config configServer.ConfigProvider) {

	if config.StoreInterval() == nil || *config.StoreInterval() <= 0 {
		return
	}

	metricStoreIntervalTicker := time.NewTicker(*config.StoreInterval())
	defer func() {
		metricStoreIntervalTicker.Stop()
	}()

	for {
		select {
		case <-metricStoreIntervalTicker.C:

			err := metricInitializer.Save()
			if err != nil {
				logger.Error("MetricInitializer", "err", err)
			}
		}
	}
}
