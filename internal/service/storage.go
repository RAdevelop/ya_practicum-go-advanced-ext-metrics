package service

/*
MetricStorage - интерфейс по работе с хранилищем метрик
скорее всего этот интерфейс надо разделить на два:
- для counter
- для gauge
*/
type MetricStorage interface {
	GaugeUpdate(name string, value float64)
	GaugeByName(name string) (float64, error)
	Gauge() map[string]float64

	CounterAdd(name string, value int64)
	CounterAccumulative() map[string]int64
	CounterAccumulativeByName(name string) (int64, error)
}
