package metric

import "context"

// Service - сервис для работы с метриками
type Service struct {
	storage Storage
}

func NewService(storage Storage) *Service {
	return &Service{
		storage: storage,
	}
}

func (ms *Service) GaugeUpdate(name string, value float64) {
	ms.storage.GaugeUpdate(name, value)
}

func (ms *Service) CounterAdd(name string, value int64) {
	ms.storage.CounterAdd(name, value)
}

func (ms *Service) CounterByNameAccumulative(name string) (int64, error) {
	return ms.storage.CounterAccumulativeByName(name)
}

func (ms *Service) GaugeByName(name string) (float64, error) {
	return ms.storage.GaugeByName(name)
}

func (ms *Service) Gauge() map[string]float64 {
	return ms.storage.Gauge()
}

func (ms *Service) CounterAccumulative() map[string]int64 {
	return ms.storage.CounterAccumulative()
}

func (ms *Service) Ping(ctx context.Context) error {
	return ms.storage.Ping(ctx)
}
