package handler

import "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"

type Handlers struct {
	Metric *Metric
}

var memStorage = service.NewMemStorage()

func New() *Handlers {
	return &Handlers{
		Metric: NewMetric(memStorage),
	}
}
