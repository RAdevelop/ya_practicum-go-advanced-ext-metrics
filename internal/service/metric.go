package service

// MetricService - сервис для работы с метриками
type MetricService struct {
	storage MetricStorage
}

func NewMetricService(storage MetricStorage) *MetricService {
	return &MetricService{
		storage: storage,
	}
}

func (ms *MetricService) GaugeUpdate(name string, value float64) {
	ms.storage.GaugeUpdate(name, value)
}

func (ms *MetricService) CounterAdd(name string, value int64) {
	ms.storage.CounterAdd(name, value)
}
