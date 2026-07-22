package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/agent"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/go-resty/resty/v2"
)

// runtimeMetrics - карта с метриками, которые будет обновлять и отправлять на сервер
var runtimeMetrics map[string]any
var PollCount int64

func main() {
	srvAddress := &serverAddress{
		host: "localhost",
		port: 8080,
	}
	_ = flag.Value(srvAddress)

	flag.Var(srvAddress, "a", `Server address pattern: "host:port without schema"`)
	rInterval := flag.Int("r", 10, `The frequency of sending metrics to the server in seconds`)
	pInterval := flag.Int("p", 2, `The frequency of metrics polling in seconds`)
	flag.Parse()

	httpClient := resty.New()
	httpClient.SetBaseURL(srvAddress.String())

	httpAgent := agent.New(httpClient)

	pollInterval := time.NewTicker(time.Duration(*pInterval) * time.Second)
	reportInterval := time.NewTicker(time.Duration(*rInterval) * time.Second)

	defer func() {
		pollInterval.Stop()
		reportInterval.Stop()
	}()

	for {
		select {
		case <-pollInterval.C: // Обновлять метрики из пакета `runtime` с заданной частотой: `pollInterval` — 2 секунды.
			PollCount++
			runtimeMetrics = collectRuntimeMetrics()
		case <-reportInterval.C: // Отправлять метрики на сервер с заданной частотой: `reportInterval` — 10 секунд.
			runtimeMetricSend(httpAgent)
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
}

func runtimeMetricSend(httpAgent *agent.HttpAgent) {
	for name, value := range runtimeMetrics {
		m := agent.MetricIn{
			Type:  models.Gauge,
			Name:  name,
			Value: fmt.Sprintf("%v", value),
		}

		metricUpdate(httpAgent, m)
	}

	metricUpdate(httpAgent, agent.MetricIn{
		Type: models.Gauge,
		Name: "RandomValue",
		Value: (func(min, max float64) string {
			rnd := min + rand.Float64()*(max-min)
			return fmt.Sprintf("%v", rnd)
		})(0, 1000),
	})

	metricUpdate(httpAgent, agent.MetricIn{
		Type:  models.Counter,
		Name:  "PollCount",
		Value: fmt.Sprintf("%v", PollCount),
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
