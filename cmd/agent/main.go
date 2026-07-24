package main

import (
	"flag"
	"io"
	"log"
	"math/rand"
	"runtime"
	"strconv"
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
	var pollCount = pollCountGet(httpAgent)

	pollInterval := time.NewTicker(time.Duration(*pInterval) * time.Second)
	reportInterval := time.NewTicker(time.Duration(*rInterval) * time.Second)

	defer func() {
		pollInterval.Stop()
		reportInterval.Stop()
	}()

	for {
		select {
		case <-pollInterval.C: // Обновлять метрики из пакета `runtime` с заданной частотой: `pollInterval` — 2 секунды.
			// - `PollCount` (тип counter) — счётчик, увеличивающийся на 1 при каждом обновлении метрики из пакета `runtime` (на каждый `pollInterval`).

			/*
						Не уверен, что правильно понял замечание.
						Эту часть задачи понял так.
						pollCount - увеличивается на 1 не когда данные на сервер отправились (там обновились), а именно по таймеру у агента.
						То есть, если:
							pollInterval = 2 сек
							reportInterval = 10 сек
						то значение у pollCount успеет увеличиться до 5 перед отправкой на сервер.
					Тогда каждые следующие 10 сек pollCount на сервере увеличивается с шагом 5

					reportInterval:	[10		20	30	40	50]
					pollInterval:	[5		10	15	20	25]
					pollCountlSum:	[5		15 	30	50	75] <- допустим тут остановили агент
					получается, что за 50 секунд времени метрики обновлялись 75 раз.
				Если запустить агента, то теперь он насчет счетчик  pollCount с 75, и далее так же с шагом 5 будет обновлять на сервере счетчик.
				Поэтому я не понял, что именно тут надо исправить, кроме как:
				"что значение количества сборов метрик, хранящееся на сервере, должно быть правильным даже при перезапуске агента"
				Это сделал.
			*/
			pollCount++
			runtimeMetrics = collectRuntimeMetrics()
		case <-reportInterval.C: // Отправлять метрики на сервер с заданной частотой: `reportInterval` — 10 секунд.
			runtimeMetricSend(httpAgent, pollCount, runtimeMetrics)
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

func pollCountGet(httpAgent *agent.HttpAgent) int64 {

	metric := agent.MetricIn{
		Type: models.Counter,
		Name: metricNamePollCount,
	}
	resp, err := httpAgent.Get(metric)

	if err != nil {
		return 0
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Printf("Error body closing for get metric: %v, err: %v\n", metric, err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error body reading for get metric: %v, err: %v\n", metric, err)
	}
	pollCount, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		return 0
	}

	return pollCount
}
func runtimeMetricSend(httpAgent *agent.HttpAgent, pollCount int64, runtimeMetrics map[string]any) {
	/*
		Если интервал времени отправки метрик на сервер будет "чаще", чем интервал времени сбора метрик, то карта с метриками может быть еще "пустой".
		Поэтому, метрики без данных не отправляем.
	*/
	if len(runtimeMetrics) == 0 {
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
