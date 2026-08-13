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

type serverSettings struct {
	ServerAddress         string
	MetricStoreInterval   uint
	MetricFileStoragePath string
	MetricRestoreFromFile bool
}

func main() {

	logMe := logger.New()
	var metricStorage = memory.NewStorage()
	var metricService = service.NewMetric(metricStorage)

	var serverConfig configServer.ConfigProvider

	configServerEnv, err := configServer.NewEnv()
	if err != nil {
		logMe.Error("error", "err", err)
		return
	}
	serverConfig = configServer.New(configServerEnv)

	srvAddress := flag.String("a", "localhost:8080", `Server address pattern: "host:port"`)
	srvStoreInterval := flag.Uint("i", 300, `интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)`)
	srvFileStoragePath := flag.String("f", "dump/metrics/iter9.json", `путь до файла, куда сохраняются текущие значения`)
	srvRestore := flag.Bool("r", true, `булево значение (true/false), определяющее, следует ли загружать ранее сохранённые значения из указанного файла при старте сервера.`)

	flag.Parse()

	servSettings := srvSettings(serverConfig, srvAddress, srvStoreInterval, srvFileStoragePath, srvRestore)

	h := handler.New(metricService, logMe)
	r := router.New(h)

	metricInitializer, err := service.NewMetricInitializer(servSettings.MetricFileStoragePath, metricService)
	if err != nil {
		logMe.Error("error", "err", err)
		return
	}

	if servSettings.MetricRestoreFromFile {
		err = metricInitializer.Load()
		if err != nil {
			logMe.Error("metricSaver Load", "err", err)
		}
	}

	go saver(metricInitializer, logMe, servSettings)

	err = http.ListenAndServe(servSettings.ServerAddress, r)
	if err != nil {
		logMe.Error("error", "err", err)
	}
}

func srvSettings(serverConfig configServer.ConfigProvider, srvAddress *string, srvStoreInterval *uint, srvFileStoragePath *string, srvRestore *bool) serverSettings {

	var settings = serverSettings{
		ServerAddress:         *srvAddress,
		MetricStoreInterval:   *srvStoreInterval,
		MetricFileStoragePath: *srvFileStoragePath,
		MetricRestoreFromFile: *srvRestore,
	}

	if serverConfig.Address() != "" {
		settings.ServerAddress = serverConfig.Address()
	}

	if serverConfig.StoreInterval() != nil {
		settings.MetricStoreInterval = *serverConfig.StoreInterval()
	}

	if serverConfig.FileStoragePath() != "" {
		settings.MetricFileStoragePath = serverConfig.FileStoragePath()
	}

	if serverConfig.Restore() != nil {
		settings.MetricRestoreFromFile = *serverConfig.Restore()
	}

	return settings
}

func saver(metricInitializer *service.MetricInitializer, logger logger.Logger, settings serverSettings) {

	metricStoreIntervalTicker := time.NewTicker(time.Duration(settings.MetricStoreInterval) * time.Second)
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
