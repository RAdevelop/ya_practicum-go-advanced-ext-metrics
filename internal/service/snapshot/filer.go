package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
)

var ErrEmptyFilePath = errors.New("file path is empty")

/*
Filer - инициализатор метрик
  - сохраняет в файл
  - загружает (инициализирует) ранее сохраненные метрики
*/
type Filer struct {
	metricService *metric.MetricService
	fileName      string
	file          *os.File
	metrics       []models.Metrics
}

func NewFiler(metricService *metric.MetricService, fileName string) (*Filer, error) {

	if fileName == "" {
		return nil, fmt.Errorf("%w: fileName = %s", ErrEmptyFilePath, fileName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fileName = filepath.Join(cwd, fileName)

	return &Filer{
		fileName:      fileName,
		metricService: metricService,
	}, nil
}

func (filer *Filer) Load() (err error) {
	defer func() {
		errClose := filer.closeFile()
		err = errors.Join(err, errClose)
	}()

	err = filer.openFile()
	if err != nil {
		return err
	}

	stat, err := filer.file.Stat()
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return nil
	}

	decoder := json.NewDecoder(filer.file)
	// Читаем открывающую скобку массива
	if _, err = decoder.Token(); err != nil {
		return err
	}
	for decoder.More() {
		modelMetric := models.Metrics{}
		err = decoder.Decode(&modelMetric)

		if err != nil {
			return err
		}

		switch modelMetric.MType {
		case models.Gauge:
			filer.metricService.GaugeUpdate(modelMetric.ID, *modelMetric.Value)
		case models.Counter:
			filer.metricService.CounterAdd(modelMetric.ID, *modelMetric.Delta)
		}
	}

	// Читаем закрывающую скобку массива
	if _, err = decoder.Token(); err != nil {
		return err
	}

	return nil
}
func (filer *Filer) Save() error {
	filer.readFromStorage()

	if len(filer.metrics) == 0 {
		return nil
	}

	filePath := filer.fileName
	tmpFile, err := filer.createTempFile(filePath)
	if err != nil {
		return err
	}

	// Записываем данные во временный файл
	encoder := json.NewEncoder(tmpFile)

	if err = encoder.Encode(filer.metrics); err != nil {
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

// readFromStorage - получаем данные метрик из источника
func (filer *Filer) readFromStorage() {
	gaugeMetrics := filer.metricService.Gauge()
	counterMetrics := filer.metricService.CounterAccumulative()

	if len(gaugeMetrics) == 0 && len(counterMetrics) == 0 {
		return
	}

	filer.metrics = make([]models.Metrics, 0, len(gaugeMetrics)+len(counterMetrics))

	if len(gaugeMetrics) > 0 {
		for id, value := range gaugeMetrics {
			modelMetric := models.Metrics{
				ID:    id,
				MType: models.Gauge,
				Value: &value,
			}

			filer.metrics = append(filer.metrics, modelMetric)
		}
	}

	if len(counterMetrics) > 0 {
		for id, value := range counterMetrics {
			modelMetric := models.Metrics{
				ID:    id,
				MType: models.Counter,
				Delta: &value,
			}

			filer.metrics = append(filer.metrics, modelMetric)
		}
	}
}

func (filer *Filer) openFile() error {
	file, err := os.OpenFile(filer.fileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	filer.file = file
	return nil
}

func (filer *Filer) closeFile() error {

	if filer.file != nil {
		err := filer.file.Close()
		filer.file = nil
		return err
	}
	return nil
}

// createTempFile - создает временный файл в той же директории
func (filer *Filer) createTempFile(filePath string) (*os.File, error) {
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
