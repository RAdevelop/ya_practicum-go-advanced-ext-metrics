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
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
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

	/*
		Понимаю, что тут что-то не так уже идет с архитектурой зависимостей.
		Надо будет ее пересмотреть, как завершу курс по Кафка (еще месяц примерно на курсе...).
		Например, metricInitializer - будет "фасадом":
			- для metricService.
			- создать интерфейс для такого менеджера (загрузка данных, сохранение данных в/из некоего источника)
		metricInitializer - имеет смысл переименовать в metricManger, и в обработчиках роутеров уже работать с ним, а не с зоопарком классов
	*/

	var metricStorage = memory.NewStorage()
	var metricService = metric.NewMetric(metricStorage)
	metricSnapshot, err := snapshot.NewFiler(metricService, serverConfig.FileStoragePath())
	if err != nil {
		logMe.Error("snapshot.NewFiler", "err", err)
		return
	}
	var metricManager = service.NewManager(metricService, metricSnapshot)

	h := handler.New(metricManager, logMe, serverConfig)
	r := router.New(h)

	if serverConfig.Restore() != nil && *serverConfig.Restore() {
		err = metricManager.MetricSnapshotLoad()
		if err != nil {
			logMe.Error("metricSaver Load", "err", err)
		}
	}

	go saver(metricManager, logMe, serverConfig)

	err = http.ListenAndServe(serverConfig.Address(), r)
	if err != nil {
		logMe.Error("error", "err", err)
	}
}

func saver(metricManager service.MetricManagementAble, logger logger.Logger, config configServer.ConfigProvider) {

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

			err := metricManager.MetricSnapshotSave()
			if err != nil {
				logger.Error("MetricInitializer", "err", err)
			}
		}
	}
}
