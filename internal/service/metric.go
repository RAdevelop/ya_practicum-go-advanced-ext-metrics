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

func (ms *MetricService) CounterByName(name string) ([]int64, error) {
	return ms.storage.CounterByName(name)
}

func (ms *MetricService) CounterByNameAccumulative(name string) (int64, error) {
	counterValues, err := ms.CounterByName(name)
	if err != nil {
		return 0, err
	}

	var sum int64
	for _, counterValue := range counterValues {
		sum += counterValue
	}

	return sum, nil
}

func (ms *MetricService) GaugeByName(name string) (float64, error) {
	return ms.storage.GaugeByName(name)
}

func (ms *MetricService) Gauge() map[string]float64 {
	return ms.storage.Gauge()
}

func (ms *MetricService) Counter() map[string][]int64 {
	return ms.storage.Counter()
}
