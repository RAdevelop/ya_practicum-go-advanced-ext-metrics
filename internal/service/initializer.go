package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

/*
MetricInitializer - инициализатор метрик
  - сохраняет в файл
  - загружает (инициализирует) ранее сохраненные метрики
*/
type MetricInitializer struct {
	fileName      string
	file          *os.File
	metricService *MetricService
	metrics       []models.Metrics
}

func NewMetricInitializer(fileName string, metricService *MetricService) (*MetricInitializer, error) {

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fileName = filepath.Join(cwd, fileName)

	return &MetricInitializer{
		fileName:      fileName,
		metricService: metricService,
	}, nil
}

func (ms *MetricInitializer) Save() (err error) {

	defer func() {
		errCloseFile := ms.closeFile()
		err = errors.Join(err, errCloseFile)
	}()

	ms.readFromStorage()
	if len(ms.metrics) == 0 {
		return nil
	}

	err = ms.openFile()
	if err != nil {
		return err
	}

	err = json.NewEncoder(ms.file).Encode(ms.metrics)
	if err != nil {
		return err
	}

	if err = ms.file.Sync(); err != nil {
		return err
	}
	return nil
}

func (ms *MetricInitializer) Load() (err error) {
	defer func() {
		err = ms.closeFile()
	}()

	err = ms.openFile()
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(ms.file)
	// Читаем открывающую скобку массива
	if _, err = decoder.Token(); err != nil {
		return err
	}
	for decoder.More() {
		metric := models.Metrics{}
		err = decoder.Decode(&metric)

		if err != nil {
			return err
		}

		switch metric.MType {
		case models.Gauge:
			ms.metricService.GaugeUpdate(metric.ID, *metric.Value)
		case models.Counter:
			ms.metricService.CounterAdd(metric.ID, *metric.Delta)
		}
	}

	// Читаем закрывающую скобку массива
	if _, err = decoder.Token(); err != nil {
		return err
	}

	return nil
}

// readFromStorage - получаем данные метрик из источника
func (ms *MetricInitializer) readFromStorage() {
	gaugeMetrics := ms.metricService.Gauge()
	counterMetrics := ms.metricService.CounterAccumulative()

	if len(gaugeMetrics) == 0 && len(counterMetrics) == 0 {
		return
	}

	ms.metrics = make([]models.Metrics, 0, len(gaugeMetrics)+len(counterMetrics))

	if len(gaugeMetrics) > 0 {
		for id, value := range gaugeMetrics {
			metric := models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			}

			ms.metrics = append(ms.metrics, metric)
		}
	}

	if len(counterMetrics) > 0 {
		for id, value := range counterMetrics {
			metric := models.Metrics{
				ID:    id,
				MType: models.Counter,
				Delta: &value,
			}

			ms.metrics = append(ms.metrics, metric)
		}
	}
}

func (ms *MetricInitializer) openFile() error {

	if !ms.fileExists() {
		err := os.MkdirAll(filepath.Dir(ms.fileName), 0755)
		if err != nil {
			return err
		}
	}

	file, err := os.OpenFile(ms.fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	ms.file = file
	return nil
}

func (ms *MetricInitializer) fileExists() bool {
	if _, err := os.Stat(ms.fileName); err == nil {
		return true
	}
	return false
}

func (ms *MetricInitializer) closeFile() error {

	if ms.file != nil {
		err := ms.file.Close()
		ms.file = nil
		return err
	}
	return nil
}
