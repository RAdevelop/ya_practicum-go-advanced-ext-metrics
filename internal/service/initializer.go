package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

var ErrEmptyFilePath = errors.New("file path is empty")

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

	if fileName == "" {
		return nil, fmt.Errorf("%w: fileName = %s", ErrEmptyFilePath, fileName)
	}

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

/*
Save - сохраняем метрики

Пояснения к: "Другой вариант -- написать обертку над хранилищем в памяти, которая будет обеспечивать работу (извлечение из/сохранение в) с файлом"

MetricInitializer - идея как раз была в этом - обертка. Я думал, написать интерфейс, чтобы MetricInitializer его реализовывал.
Как раз для того, чтобы можно было менять при необходимости работу с файлами, на другое хранилище (БД, внешний сервис и тп).
Причем такой, который реализовывал бы io.Reader, io.Writer

Не стал делать, чтобы не увлеичивать кодовую базу. И как читал по Go - создавать интерфейсы лучше тогда, когда действиетльно нужно по задаче :)
И не надо "гадать" - пригодится или нет. Надо было мне сразу такой комментарий оставить.
MetricInitializer - тут скорее неудачно имя выбрал. Из опыта знаю, что в разработке самое сложное: это выбор имен и задачт кэширования :)
*/
func (ms *MetricInitializer) Save() (err error) {
	ms.readFromStorage()

	if len(ms.metrics) == 0 {
		return nil
	}

	filePath := ms.fileName
	tmpFile, err := ms.createTempFile(filePath)
	if err != nil {
		return err
	}

	// Записываем данные во временный файл
	encoder := json.NewEncoder(tmpFile)

	if err = encoder.Encode(ms.metrics); err != nil {
		return err
	}

	if err = tmpFile.Sync(); err != nil {
		return err
	}

	if err = tmpFile.Close(); err != nil {
		return err
	}

	// Атомарная замена
	if err = os.Rename(tmpFile.Name(), filePath); err != nil {
		return err
	}

	return nil
}

func (ms *MetricInitializer) Load() (err error) {
	defer func() {
		errClose := ms.closeFile()
		err = errors.Join(err, errClose)
	}()

	err = ms.openFile()
	if err != nil {
		return err
	}

	stat, err := ms.file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
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
	file, err := os.OpenFile(ms.fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	ms.file = file
	return nil
}

func (ms *MetricInitializer) closeFile() error {

	if ms.file != nil {
		err := ms.file.Close()
		ms.file = nil
		return err
	}
	return nil
}

// createTempFile - создает временный файл в той же директории
func (ms *MetricInitializer) createTempFile(filePath string) (*os.File, error) {
	// Получаем директорию файла
	dir := filepath.Dir(filePath)

	// Создаем директорию, если её нет
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Создаем временный файл с префиксом
	tmpFile, err := os.CreateTemp(dir, "metrics_*.tmp")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	return tmpFile, nil
}
