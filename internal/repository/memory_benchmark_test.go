package repository

import (
	"context"
	"fmt"
	"testing"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

// gaugeMetric — конструктор gauge-метрики для бенчмарков.
func gaugeMetric(id string, value float64) models.Metrics {
	v := value
	return models.Metrics{ID: id, MType: models.Gauge, Value: &v}
}

// counterMetric — конструктор counter-метрики для бенчмарков.
func counterMetric(id string, delta int64) models.Metrics {
	d := delta
	return models.Metrics{ID: id, MType: models.Counter, Delta: &d}
}

// makeBatch — формирует пачку метрик: первые half — gauge, остальные — counter.
// uniqueIDs=true даёт уникальные имена (заполняем map), false — одни и те же ID
// (перезапись gauge и накопление counter, как в реальном потоке метрик).
func makeBatch(n int, uniqueIDs bool) []models.Metrics {
	metrics := make([]models.Metrics, 0, n)
	for i := 0; i < n; i++ {
		id := "PollCount"
		if uniqueIDs {
			id = fmt.Sprintf("metric_%d", i)
		}
		if i%2 == 0 {
			metrics = append(metrics, gaugeMetric(id, float64(i)))
		} else {
			metrics = append(metrics, counterMetric(id, 1))
		}
	}
	return metrics
}

// BenchmarkMemStorage_UpdateBatch — производительность батч-обновления
// при разном размере пачки и уникальности ID.
func BenchmarkMemStorage_UpdateBatch(b *testing.B) {
	for _, n := range []int{1, 10, 100, 1000} {
		for _, unique := range []bool{true, false} {
			name := fmt.Sprintf("batch_%d/unique_ids_%v", n, unique)
			metrics := makeBatch(n, unique)

			b.Run(name, func(b *testing.B) {
				storage := NewMemory()
				ctx := context.Background()

				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if _, err := storage.UpdateBatch(ctx, metrics); err != nil {
						b.Fatalf("UpdateBatch: %v", err)
					}
				}
			})
		}
	}
}

// BenchmarkMemStorage_UpdateBatchParallel — конкурентное батч-обновление.
// Показывает влияние sync.RWMutex на пропускную способность.
func BenchmarkMemStorage_UpdateBatchParallel(b *testing.B) {
	for _, n := range []int{1, 100} {
		metrics := makeBatch(n, false)

		b.Run(fmt.Sprintf("batch_%d", n), func(b *testing.B) {
			storage := NewMemory()
			ctx := context.Background()

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if _, err := storage.UpdateBatch(ctx, metrics); err != nil {
						b.Fatalf("UpdateBatch: %v", err)
					}
				}
			})
		})
	}
}

// BenchmarkMemStorage_UpdateBatchCounterAccumulate — худший для накопления случай:
// все батчи инкрементируют один и тот же counter, словарь остаётся крошечным.
func BenchmarkMemStorage_UpdateBatchCounterAccumulate(b *testing.B) {
	storage := NewMemory()
	ctx := context.Background()
	metrics := []models.Metrics{counterMetric("PollCount", 1)}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := storage.UpdateBatch(ctx, metrics); err != nil {
			b.Fatalf("UpdateBatch: %v", err)
		}
	}
}

// BenchmarkMemStorage_UpdateBatchHeavyDedup — кейс с кучей дублей:
// baseStorage.AddValuesForDeduplicateMetrics должен их схлопнуть,
// а MemStorage — обработать уже короткую пачку.
func BenchmarkMemStorage_UpdateBatchHeavyDedup(b *testing.B) {
	const n = 10_000 // количество метрик с одинаковым ID = "PollCount"

	// Увеличиваем количество дублей: каждый ID встречается много раз
	// (в makeBatch уже false, но здесь мы делаем ещё более экстремальный кейс)
	metrics := makeBatch(n, false)

	storage := NewMemory()
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := storage.UpdateBatch(ctx, metrics); err != nil {
			b.Fatalf("UpdateBatch: %v", err)
		}
	}
}

// BenchmarkMemStorage_UpdateBatchErrorPath — цена раннего возврата с ошибкой:
// метрика с пустым ID прерывает цикл на первом элементе.
func BenchmarkMemStorage_UpdateBatchErrorPath(b *testing.B) {
	storage := NewMemory()
	ctx := context.Background()
	metrics := makeBatch(1000, true)
	metrics[0] = models.Metrics{ID: "", MType: models.Gauge}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := storage.UpdateBatch(ctx, metrics); err == nil {
			b.Fatal("expected error, got nil")
		}
	}
}
