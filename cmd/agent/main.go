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

type agentFlags struct {
	IntervalReport *uint
	IntervalPoll   *uint
	SingKey        *string
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

	agFlags := &agentFlags{}

	flag.Var(srvAddress, "a", `Server address pattern: "host:port without schema"`)
	agFlags.IntervalReport = flag.Uint("r", 10, `The frequency of sending metrics to the server in seconds`)
	agFlags.IntervalPoll = flag.Uint("p", 2, `The frequency of metrics polling in seconds`)
	agFlags.SingKey = flag.String("k", "", `The key used to sign the request`)
	flag.Parse()

	configAgentEnv, err := configAgent.NewEnv()
	if err != nil {
		logApp.Error("error", fmt.Errorf("error creating configAgent environment variable: %w", err))
		return
	}

	agentConfig := configAgent.New(configAgentEnv)

	agentConfigUpdate(agentConfig, srvAddress.String(), agFlags)

	httpClient := resty.New()
	httpClient.SetBaseURL("http://" + agentConfig.Address())

	httpAgent := agent.New(httpClient, agentConfig)
	var pollCount = int64(0)

	pollInterval := time.NewTicker(time.Duration(agentConfig.PollInterval()) * time.Second)
	reportInterval := time.NewTicker(time.Duration(agentConfig.ReportInterval()) * time.Second)

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

func agentConfigUpdate(agentConfig *configAgent.Config, srvAddress string, agFlags *agentFlags) {

	if agentConfig == nil {
		return
	}

	if srvAddress != "" {
		agentConfig.AddressSet(srvAddress)
	}

	if agFlags == nil {
		return
	}

	if agFlags.IntervalReport != nil {
		agentConfig.ReportIntervalSet(*agFlags.IntervalReport)
	}

	if agFlags.IntervalPoll != nil {
		agentConfig.PollIntervalSet(*agFlags.IntervalPoll)
	}

	if agFlags.SingKey != nil {
		agentConfig.SignKeySet(*agFlags.SingKey)
	}
}

func metricUpdateBatch(ctx context.Context, httpAgent *agent.HTTPAgent, metrics []models.Metrics) (err error) {

	resp, err := httpAgent.Updates(ctx, metrics)
	defer func() {
		if resp != nil && resp.Body != nil {
			closeErr := resp.Body.Close()
			err = errors.Join(err, closeErr)
		}
	}()

	return handleUpdateResponse(resp, err, metrics)
}

func handleUpdateResponse(resp *http.Response, errResp error, metric any) (err error) {
	if errResp != nil {
		err = fmt.Errorf("error updating metric: %v, err: %w", metric, errResp)
		return
	}

	// io.Discard выступает в качестве приёмника ненужных данных.
	// Ведь надо всегда считывать тело сообщения, даже если оно не нужно?!
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		err = fmt.Errorf("error body reading for updating metric: %v, err: %w", metric, err)
		return
	}
	return nil
}

func runtimeMetricSend(ctx context.Context, logApp logger.Logger, httpAgent *agent.HTTPAgent, pollCount int64, runtimeMetrics []models.Metrics) {
	/*
		Если интервал времени отправки метрик на сервер будет "чаще", чем интервал времени сбора метрик, то карта с метриками может быть еще "пустой".
		Поэтому, метрики без данных не отправляем.
	*/
	if len(runtimeMetrics) == 0 || pollCount == 0 {
		return
	}

	err := metricUpdateBatch(ctx, httpAgent, runtimeMetrics)
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
