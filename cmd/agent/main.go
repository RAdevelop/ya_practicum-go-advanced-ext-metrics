package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
	configAgent "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

const metricNamePollCount = "PollCount"

func main() {
	// runtimeMetrics - карта с метриками, которые будем обновлять и отправлять на сервер
	var runtimeMetrics map[string]any

	srvAddress := &agent.ServerAddress{
		Host: "localhost",
		Port: 8080,
	}
	_ = flag.Value(srvAddress)

	flag.Var(srvAddress, "a", `Server address pattern: "host:port without schema"`)
	rInterval := flag.Uint("r", 10, `The frequency of sending metrics to the server in seconds`)
	pInterval := flag.Uint("p", 2, `The frequency of metrics polling in seconds`)
	flag.Parse()

	var agentConfig configAgent.ConfigProvider

	configAgentEnv, err := configAgent.NewEnv()
	if err != nil {
		log.Fatal(err)
	}
	agentConfig = configAgent.New(configAgentEnv)

	serverAddress := srvAddress.String()
	intervalReport := *rInterval
	intervalPoll := *pInterval

	if agentConfig.Address() != "" {
		serverAddress = agentConfig.Address()
	}

	if agentConfig.ReportInterval() > 0 {
		intervalReport = agentConfig.ReportInterval()
	}
	if agentConfig.PollInterval() > 0 {
		intervalPoll = agentConfig.PollInterval()
	}

	httpClient := resty.New()
	httpClient.SetBaseURL("http://" + serverAddress)

	httpAgent := agent.New(httpClient)
	var pollCount = int64(0)

	pollInterval := time.NewTicker(time.Duration(intervalPoll) * time.Second)
	reportInterval := time.NewTicker(time.Duration(intervalReport) * time.Second)

	defer func() {
		pollInterval.Stop()
		reportInterval.Stop()
	}()

	for {
		select {
		case <-pollInterval.C: // Обновлять метрики из пакета `runtime` с заданной частотой: `pollInterval` — 2 секунды.
			runtimeMetrics = collectRuntimeMetrics()
			pollCount++
		case <-reportInterval.C: // Отправлять метрики на сервер с заданной частотой: `reportInterval` — 10 секунд.
			runtimeMetricSend(httpAgent, pollCount, runtimeMetrics)
			pollCount = 0
		}
	}
}

func metricUpdate(httpAgent *agent.HttpAgent, metric models.Metrics) (err error) {

	resp, err := httpAgent.Update(metric)
	if err != nil {
		err = fmt.Errorf("Error updating metric: %v\n err: %w\n", metric, err)
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; close body error: %w", err, closeErr)
			} else {
				err = fmt.Errorf("close body error: %w", closeErr)
			}
		}
	}()

	// io.Discard выступает в качестве приёмника ненужных данных.
	// Ведь надо всегда считывать тело сообщения, даже если оно не нужно?!
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		return fmt.Errorf("error body reading for updating metric: %v, err: %w\n", metric, err)
	}
	return nil
}

func runtimeMetricSend(httpAgent *agent.HttpAgent, pollCount int64, runtimeMetrics map[string]any) {
	/*
		Если интервал времени отправки метрик на сервер будет "чаще", чем интервал времени сбора метрик, то карта с метриками может быть еще "пустой".
		Поэтому, метрики без данных не отправляем.
	*/
	if len(runtimeMetrics) == 0 || pollCount == 0 {
		return
	}

	var err error
	for name, value := range runtimeMetrics {
		v, errConvert := converter.ToFloat64(value)
		if errConvert != nil {
			continue
		}
		m := models.Metrics{
			MType: models.Gauge,
			ID:    name,
			Value: new(v),
		}

		err = metricUpdate(httpAgent, m)
		if err != nil {
			log.Printf("Error updating metric: %v, err: %v\n", m, err)
		}
	}

	m := models.Metrics{
		MType: models.Gauge,
		ID:    "RandomValue",
		Value: (func(min, max float64) *float64 {
			rnd := min + rand.Float64()*(max-min)
			return new(rnd)
		})(0, 1000),
	}

	err = metricUpdate(httpAgent, m)
	if err != nil {
		log.Printf("Error updating metric: %v, err: %v\n", m, err)
	}
	m = models.Metrics{
		MType: models.Counter,
		ID:    metricNamePollCount,
		Delta: &pollCount,
	}
	err = metricUpdate(httpAgent, m)
	if err != nil {
		log.Printf("Error updating metric: %v, err: %v\n", m, err)
	}
}

func collectRuntimeMetrics() map[string]any {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	return map[string]any{
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
}
