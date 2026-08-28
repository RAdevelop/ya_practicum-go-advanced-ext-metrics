package main

import (
	"context"
	"flag"
	"net/http"
	"time"

	configDB "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/db"
	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/database"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
)

type serverFlags struct {
	address          *string
	storeInterval    *uint
	fileStoragePath  *string
	restore          *bool
	dbDSN            *string
	useMemoryStorage bool
}

func main() {

	logApp := logger.New()

	configServerEnv, err := configServer.NewEnv()
	if err != nil {
		logApp.Error("error", "err", err)
		return
	}

	configDBEnv, err := configDB.NewEnv()
	if err != nil {
		logApp.Error("configDBEnv", "err", err)
		return
	}

	srvFlags := &serverFlags{}

	serverConfig := configServer.New(configServerEnv)
	dbConfig := configDB.New(configDBEnv)

	srvFlags.address = flag.String("a", "localhost:8080", `Server address pattern: "host:port"`)
	srvFlags.storeInterval = flag.Uint("i", 300, `интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)`)
	srvFlags.fileStoragePath = flag.String("f", "dump/metrics/iter9.json", `путь до файла, куда сохраняются текущие значения`)
	srvFlags.restore = flag.Bool("r", true, `булево значение (true/false), определяющее, следует ли загружать ранее сохранённые значения из указанного файла при старте сервера.`)
	srvFlags.dbDSN = flag.String("d", "", `Строка с адресом подключения к БД`)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverConfigUpdateByFlags(serverConfig, srvFlags)

	if srvFlags.dbDSN != nil && *srvFlags.dbDSN != "" {
		dbConfig.DSNSet(*srvFlags.dbDSN)
	}

	var metricStorage metric.Storage

	if dbConfig.DSN() == "" {
		metricStorage = memory.NewStorage()
		srvFlags.useMemoryStorage = true
	} else {
		//Из задания на самом деле не понятно точно, допустим сохранять в файл не надо, но вот восстанавливать из файла надо или нет?
		//serverConfig.RestoreSet(new(false))
		//serverConfig.StoreIntervalSet(srvFlags.storeInterval)
		db, err := database.NewDB(ctx, dbConfig)
		if err != nil {
			srvFlags.useMemoryStorage = true
			logApp.Error("db", "err", err)
		} else {
			defer db.Close()
			srvFlags.useMemoryStorage = false
		}

		metricStorage = database.NewStorage(db)
	}

	metricSnapshot, err := snapshot.NewFiler(metricStorage, serverConfig.FileStoragePath())
	if err != nil {
		logApp.Error("snapshot.NewFiler", "err", err)
		return
	}
	var metricManager = service.NewManager(metricStorage, metricSnapshot)

	h := handler.New(metricManager, logApp, serverConfig)
	r := router.New(h)

	if srvFlags.useMemoryStorage {
		metricSnapshotTask(ctx, metricManager, logApp, serverConfig)
	}

	err = http.ListenAndServe(serverConfig.Address(), r)
	if err != nil {
		logApp.Error("ListenAndServe", "err", err)
	}
}

func metricSnapshotTask(ctx context.Context, metricManager service.MetricManagementAble, logger logger.Logger, config configServer.ConfigProvider) {

	if config.Restore() != nil && *config.Restore() {
		err := metricManager.MetricSnapshotLoad(ctx)
		if err != nil {
			logger.Error("metricManager MetricSnapshotLoad", "err", err)
		}
	}

	go saver(ctx, metricManager, logger, config)
}

func saver(ctx context.Context, metricManager service.MetricManagementAble, logger logger.Logger, config configServer.ConfigProvider) {

	if config.StoreInterval() == nil || *config.StoreInterval() <= 0 {
		return
	}

	metricStoreIntervalTicker := time.NewTicker(*config.StoreInterval())
	defer func() {
		metricStoreIntervalTicker.Stop()
	}()

	for range metricStoreIntervalTicker.C {
		err := metricManager.MetricSnapshotSave(ctx)
		if err != nil {
			logger.Error("MetricInitializer", "err", err)
		}
	}
}

func serverConfigUpdateByFlags(serverConfig *configServer.Config, srvFlags *serverFlags) {
	if serverConfig.Address() == "" {
		serverConfig.AddressSet(*srvFlags.address)
	}

	if serverConfig.StoreInterval() == nil {
		serverConfig.StoreIntervalSet(srvFlags.storeInterval)
	}
	if serverConfig.FileStoragePath() == "" {
		serverConfig.FileStoragePathSet(*srvFlags.fileStoragePath)
	}

	if serverConfig.Restore() == nil {
		serverConfig.RestoreSet(srvFlags.restore)
	}
}
