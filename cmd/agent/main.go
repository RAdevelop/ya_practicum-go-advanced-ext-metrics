package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
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
	rInterval := flag.Int("r", 10, `The frequency of sending metrics to the server in seconds`)
	pInterval := flag.Int("p", 2, `The frequency of metrics polling in seconds`)
	flag.Parse()

	httpClient := resty.New()
	httpClient.SetBaseURL(srvAddress.String())

	httpAgent := agent.New(httpClient)
	var pollCount = int64(0)

	pollInterval := time.NewTicker(time.Duration(*pInterval) * time.Second)
	reportInterval := time.NewTicker(time.Duration(*rInterval) * time.Second)

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

func metricUpdate(httpAgent *agent.HttpAgent, metric agent.MetricIn) {

	resp, err := httpAgent.Update(metric)
	if err != nil {
		log.Printf("Error updating metric: %v\n err: %v\n", metric, err)
		log.Printf("Error type: %T\n", err)
		return
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Printf("Error body closing for updating metric: %v, err: %v\n", metric, err)
		}
	}()

	// io.Discard выступает в качестве приёмника ненужных данных.
	// Ведь надо всегда считывать тело сообщения, даже если оно не нужно?!
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		log.Printf("Error body reading for updating metric: %v, err: %v\n", metric, err)
	}
}

func runtimeMetricSend(httpAgent *agent.HttpAgent, pollCount int64, runtimeMetrics map[string]any) {
	/*
		Если интервал времени отправки метрик на сервер будет "чаще", чем интервал времени сбора метрик, то карта с метриками может быть еще "пустой".
		Поэтому, метрики без данных не отправляем.
	*/
	if len(runtimeMetrics) == 0 || pollCount == 0 {
		return
	}

	for name, value := range runtimeMetrics {
		m := agent.MetricIn{
			Type:  models.Gauge,
			Name:  name,
			Value: converter.NumericToString(value),
		}

		metricUpdate(httpAgent, m)
	}

	metricUpdate(httpAgent, agent.MetricIn{
		Type: models.Gauge,
		Name: "RandomValue",
		Value: (func(min, max float64) string {
			rnd := min + rand.Float64()*(max-min)
			return converter.NumericToString(rnd)
		})(0, 1000),
	})

	metricUpdate(httpAgent, agent.MetricIn{
		Type:  models.Counter,
		Name:  metricNamePollCount,
		Value: converter.NumericToString(pollCount),
	})
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
