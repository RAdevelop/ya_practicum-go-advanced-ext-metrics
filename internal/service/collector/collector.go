package collector

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// MetricsCollector — потокобезопасный накопитель метрик.
//
// Хранит не только сами метрики, но и локальный счётчик опросов pollCount.
// Счётчик сбрасывается в момент, когда накопленный батч забирается
// report-воркером, — тогда на сервер уходит дельта, равная числу
// опросов с прошлой отправки, а сервер накапливает её у себя.
type MetricsCollector struct {
	logApp    logger.Logger
	mu        sync.Mutex
	metrics   []models.Metrics
	pollCount int64
}

func New(logApp logger.Logger) *MetricsCollector {
	return &MetricsCollector{
		logApp: logApp,
	}
}

func (mc *MetricsCollector) add(metrics ...models.Metrics) {
	mc.mu.Lock()
	mc.metrics = append(mc.metrics, metrics...)
	mc.mu.Unlock()
}

// incPollCount увеличивает счётчик опросов и возвращает новое значение.
func (mc *MetricsCollector) incPollCount() int64 {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.pollCount++
	return mc.pollCount
}

// drainAndReset возвращает накопленные метрики и текущее значение pollCount,
// после чего очищает буфер и сбрасывает счётчик.
func (mc *MetricsCollector) drainAndReset() ([]models.Metrics, int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metrics := mc.metrics
	pollCount := mc.pollCount

	mc.metrics = nil
	mc.pollCount = 0

	return metrics, pollCount
}

// collectRuntimeMetrics собирает gauge-метрики из пакета runtime.
// PollCount сюда не входит — он добавляется в report-воркере.
func (mc *MetricsCollector) collectRuntimeMetrics() []models.Metrics {
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

	metrics := make([]models.Metrics, 0, len(runtimeMetrics)+1)

	for name, value := range runtimeMetrics {
		v, errConvert := converter.ToFloat64(value)
		if errConvert != nil {
			continue
		}
		metrics = append(metrics, models.Metrics{
			MType: models.Gauge,
			ID:    name,
			Value: &v,
		})
	}

	rnd := rand.Float64() * 1000
	metrics = append(metrics, models.Metrics{
		MType: models.Gauge,
		ID:    "RandomValue",
		Value: &rnd,
	})

	return metrics
}

// collectPsutilMetrics — дополнительные gauge-метрики через gopsutil.
func (mc *MetricsCollector) collectPsutilMetrics(numCPU int) ([]models.Metrics, error) {
	metrics := make([]models.Metrics, 0, 2+numCPU)

	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("mem.VirtualMemory: %w", err)
	}

	total := float64(vm.Total)
	free := float64(vm.Free)
	metrics = append(metrics,
		models.Metrics{MType: models.Gauge, ID: "TotalMemory", Value: &total},
		models.Metrics{MType: models.Gauge, ID: "FreeMemory", Value: &free},
	)

	// percpu=true → слайс длиной numCPU, по одному значению на ядро.
	percents, err := cpu.Percent(0, true)
	if err != nil {
		return nil, fmt.Errorf("cpu.Percent: %w", err)
	}

	for i := 0; i < numCPU && i < len(percents); i++ {
		v := percents[i]
		metrics = append(metrics, models.Metrics{
			MType: models.Gauge,
			ID:    fmt.Sprintf("CPUutilization%d", i+1),
			Value: &v,
		})
	}

	return metrics, nil
}

// RunPollWorker — раз в interval собирает runtime-метрики.
func (mc *MetricsCollector) RunPollWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.incPollCount()
			metrics := mc.collectRuntimeMetrics()
			mc.add(metrics...)
		}
	}
}

// RunPsutilWorker — раз в interval собирает метрики через gopsutil.
func (mc *MetricsCollector) RunPsutilWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Число CPU узнаём один раз при старте — оно не меняется.
	numCPU, err := cpu.Counts(true)
	if err != nil || numCPU < 1 {
		numCPU = 1
	}
	_, _ = cpu.Percent(0, true)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, err := mc.collectPsutilMetrics(numCPU)
			if err != nil {
				mc.logApp.Error("collectPsutilMetrics", "err", err)
				continue
			}
			mc.incPollCount()
			mc.add(metrics...)
		}
	}
}

// RunReportWorker — раз в interval забирает снапшот и кладёт его в jobs.
//
// PollCount добавляется сюда: значение Delta = числу опросов с прошлой
// отправки. Так сервер накапливает counter корректно.
func (mc *MetricsCollector) RunReportWorker(ctx context.Context, interval time.Duration, jobs chan<- []models.Metrics) {
	defer close(jobs)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics, pollCount := mc.drainAndReset()

			if len(metrics) == 0 && pollCount == 0 {
				continue
			}

			metrics = append(metrics, models.Metrics{
				MType: models.Counter,
				ID:    "PollCount",
				Delta: &pollCount,
			})

			select {
			case jobs <- metrics:
			case <-ctx.Done():
				return
			}
		}
	}
}
