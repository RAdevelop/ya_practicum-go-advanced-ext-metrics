package memory

import "errors"

var ErrNotFoundName = errors.New("metric not found by name")

// MemStorage - хранилище метрик в памяти
type MemStorage struct {
	gauge   map[string]float64
	counter map[string][]int64
}

/*
NewMemStorage - конструктор для структуры хранения метрик
TODO: в будущем, скорее всего стоит добавить размер для инициализации карт:
  - ms.gauge = make(map[string]float64)
  - ms.counter = make(map[string][]int64)
    Чтобы при можно было изначально ориентироваться на известный размер (кол-во метрик) для оптимизации работы с памятью при добавлении метрик:
    func NewMemStorage(gaugeSize int64, counterSize int64) *MemStorage{...}

TODO наверное, надо будет добавить mutex для обработки ситуации с гонкой данных.
*/
func NewMemStorage() *MemStorage {

	memStorage := &MemStorage{}
	memStorage.gaugeInit()
	memStorage.counterInit()

	return memStorage
}

func (ms *MemStorage) GaugeUpdate(name string, value float64) {
	ms.gaugeInit()
	ms.gauge[name] = value
}

func (ms *MemStorage) GaugeByName(name string) (float64, error) {

	ms.gaugeInit()
	if value, ok := ms.gauge[name]; ok {
		return value, nil
	}

	return 0, ErrNotFoundName
}

func (ms *MemStorage) Gauge() map[string]float64 {
	ms.gaugeInit()
	return ms.gauge
}

func (ms *MemStorage) GaugeSize() int {
	return len(ms.gauge)
}

func (ms *MemStorage) CounterAdd(name string, value int64) {

	ms.counterInit()
	if _, ok := ms.counter[name]; !ok {
		ms.counter[name] = make([]int64, 0)
	}

	ms.counter[name] = append(ms.counter[name], value)
}

func (ms *MemStorage) CounterByName(name string) ([]int64, error) {

	ms.counterInit()
	if value, ok := ms.counter[name]; ok {
		return value, nil
	}

	return nil, ErrNotFoundName
}

func (ms *MemStorage) Counter() map[string][]int64 {
	ms.counterInit()
	return ms.counter
}

func (ms *MemStorage) CounterSize() int {
	ms.counterInit()
	return len(ms.counter)
}

func (ms *MemStorage) CounterSizeByName(name string) int {

	ms.counterInit()

	if _, ok := ms.counter[name]; !ok {
		return 0
	}

	return len(ms.counter[name])
}

func (ms *MemStorage) counterInit() {
	if ms.counter == nil {
		ms.counter = make(map[string][]int64)
	}
}

func (ms *MemStorage) gaugeInit() {
	if ms.gauge == nil {
		ms.gauge = make(map[string]float64)
	}
}
