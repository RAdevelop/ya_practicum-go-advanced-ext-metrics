package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
	configAgent "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

type agentSettings struct {
	ServerAddress  string
	IntervalReport uint
	IntervalPoll   uint
}

func main() {
	// runtimeMetrics - карта с метриками, которые будем обновлять и отправлять на сервер
	var runtimeMetrics []models.Metrics

	logApp := logger.New()

	srvAddress := &agent.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}
	_ = flag.Value(srvAddress)

	flag.Var(srvAddress, "a", `Server address pattern: "host:port without schema"`)
	rInterval := flag.Uint("r", 10, `The frequency of sending metrics to the server in seconds`)
	pInterval := flag.Uint("p", 2, `The frequency of metrics polling in seconds`)
	flag.Parse()

	configAgentEnv, err := configAgent.NewEnv()
	if err != nil {
		logApp.Error("error", fmt.Errorf("error creating configAgent environment variable: %w", err))
		return
	}

	agentConfig := configAgent.New(configAgentEnv)

	agSettings := settings(agentConfig, srvAddress.String(), rInterval, pInterval)

	httpClient := resty.New()
	httpClient.SetBaseURL("http://" + agSettings.ServerAddress)

	httpAgent := agent.New(httpClient)
	var pollCount = int64(0)

	pollInterval := time.NewTicker(time.Duration(agSettings.IntervalPoll) * time.Second)
	reportInterval := time.NewTicker(time.Duration(agSettings.IntervalReport) * time.Second)

	ctx, cancel := context.WithCancel(context.Background())

	defer func() {
		pollInterval.Stop()
		reportInterval.Stop()
		cancel()
	}()

	for {
		select {
		case <-pollInterval.C: // Обновлять метрики из пакета `runtime` с заданной частотой: `pollInterval` — 2 секунды.
			pollCount++
			runtimeMetrics = collectRuntimeMetrics(pollCount)
		case <-reportInterval.C: // Отправлять метрики на сервер с заданной частотой: `reportInterval` — 10 секунд.
			runtimeMetricSend(ctx, logApp, httpAgent, pollCount, runtimeMetrics)
			pollCount = 0
		}
	}
}

func settings(agentConfig configAgent.ConfigProvider, srvAddress string, intervalReport *uint, intervalPoll *uint) agentSettings {

	cfg := agentSettings{
		ServerAddress:  srvAddress,
		IntervalReport: *intervalReport,
		IntervalPoll:   *intervalPoll,
	}

	if agentConfig.Address() != "" {
		cfg.ServerAddress = agentConfig.Address()
	}

	if agentConfig.ReportInterval() > 0 {
		cfg.IntervalReport = agentConfig.ReportInterval()
	}
	if agentConfig.PollInterval() > 0 {
		cfg.IntervalPoll = agentConfig.PollInterval()
	}

	return cfg
}

/*
```

	В runtimeMetricSend метрики отправляются дважды: сначала по одной в цикле через metricUpdate, а затем всем списком через metricUpdateBatch. Итого за один тик агент делает N + 1 HTTP-запросов вместо одного -- это лишняя нагрузка и на агент, и на сервер.

	Цикл с единичными отправками стоит убрать полностью и оставить только metricUpdateBatch. Вместе с этим можно удалить и саму функцию metricUpdate, и метод Update у HttpAgent -- они больше не нужны агенту. Эндпоинт POST /update на сервере при этом никуда не девается: он остаётся для обратной совместимости с другими клиентами, просто сам агент перестаёт им пользоваться

```

В цикле через metricUpdate - да, я это видел. Я это и убирал изначально, пока не стал запускать тесты от практикума перед пушем.
Тест инкремента №7 (TestIteration7) - падает, если закомментировать цикл отправки через metricUpdate (или в методе metricUpdate вернуть nil) :)
Поэтому и оставил metricUpdate в цикле.

Я подумал, что специально сделано, чтобы в следующих инкрементах обрабатывать ситуации с конкурентными запросами :) В данном случае, наверное, сложно назвать это конкурентностью. Но все же. Внутренности сами тестов практикума я не смотрел все еще, как и рекомендовали наставники.
*/
func metricUpdate(ctx context.Context, httpAgent *agent.HttpAgent, metric models.Metrics) error {

	//return nil //если не вызвать метод, то тест TestIteration7 падает
	resp, err := httpAgent.Update(ctx, metric)
	return handleUpdateResponse(resp, err, metric)
}

func metricUpdateBatch(ctx context.Context, httpAgent *agent.HttpAgent, metrics []models.Metrics) error {

	resp, err := httpAgent.Updates(ctx, metrics)
	return handleUpdateResponse(resp, err, metrics)
}

func handleUpdateResponse(resp *http.Response, errResp error, metric any) (err error) {
	if errResp != nil {
		err = fmt.Errorf("error updating metric: %v, err: %w", metric, errResp)
		return
	}
	defer func() {
		closeErr := resp.Body.Close()
		err = errors.Join(err, closeErr)
	}()

	// io.Discard выступает в качестве приёмника ненужных данных.
	// Ведь надо всегда считывать тело сообщения, даже если оно не нужно?!
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		err = fmt.Errorf("error body reading for updating metric: %v, err: %w", metric, err)
		return
	}
	return nil
}

func runtimeMetricSend(ctx context.Context, logApp logger.Logger, httpAgent *agent.HttpAgent, pollCount int64, runtimeMetrics []models.Metrics) {
	/*
		Если интервал времени отправки метрик на сервер будет "чаще", чем интервал времени сбора метрик, то карта с метриками может быть еще "пустой".
		Поэтому, метрики без данных не отправляем.
	*/
	if len(runtimeMetrics) == 0 || pollCount == 0 {
		return
	}

	var err error

	for _, metric := range runtimeMetrics {
		err = metricUpdate(ctx, httpAgent, metric)
		if err != nil {
			logApp.Error("metricUpdate", "err", err)
		}
	}

	err = metricUpdateBatch(ctx, httpAgent, runtimeMetrics)
	if err != nil {
		logApp.Error("metricUpdateBatch", "err", err)
	}
}

func collectRuntimeMetrics(pollCount int64) []models.Metrics {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	runtimeMetrics := map[string]any{
		"Alloc":         ms.Alloc,
		"BuckHashSys":   ms.BuckHashSys,
		"Frees":         ms.Frees,
		"GCCPUFraction": ms.GCCPUFraction,
		"GCSys":         ms.GCSys,
		"HeapAlloc":     ms.HeapAlloc,
		"HeapIdle":      ms.HeapIdle,
		"HeapInuse":     ms.HeapInuse,
		"HeapObjects":   ms.HeapObjects,
		"HeapReleased":  ms.HeapReleased,
		"HeapSys":       ms.HeapSys,
		"LastGC":        ms.LastGC,
		"Lookups":       ms.Lookups,
		"MCacheInuse":   ms.MCacheInuse,
		"MCacheSys":     ms.MCacheSys,
		"MSpanInuse":    ms.MSpanInuse,
		"MSpanSys":      ms.MSpanSys,
		"Mallocs":       ms.Mallocs,
		"NextGC":        ms.NextGC,
		"NumForcedGC":   ms.NumForcedGC,
		"NumGC":         ms.NumGC,
		"OtherSys":      ms.OtherSys,
		"PauseTotalNs":  ms.PauseTotalNs,
		"StackInuse":    ms.StackInuse,
		"StackSys":      ms.StackSys,
		"Sys":           ms.Sys,
		"TotalAlloc":    ms.TotalAlloc,
	}

	metrics := make([]models.Metrics, 0, len(runtimeMetrics)+2)

	for name, value := range runtimeMetrics {
		v, errConvert := converter.ToFloat64(value)
		if errConvert != nil {
			continue
		}
		runtimeMetric := models.Metrics{
			MType: models.Gauge,
			ID:    name,
			Value: &v,
		}

		metrics = append(metrics, runtimeMetric)
	}

	mRandomValue := models.Metrics{
		MType: models.Gauge,
		ID:    "RandomValue",
		Value: (func(min, max float64) *float64 {
			rnd := min + rand.Float64()*(max-min)
			return &rnd
		})(0, 1000),
	}
	metrics = append(metrics, mRandomValue)

	mPollCount := models.Metrics{
		MType: models.Counter,
		ID:    "PollCount",
		Delta: &pollCount,
	}
	metrics = append(metrics, mPollCount)

	return metrics
}
