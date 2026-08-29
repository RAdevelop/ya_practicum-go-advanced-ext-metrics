package snapshot

import (
	"context"
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
	storage  metric.Storage
	fileName string
	file     *os.File
	metrics  []models.Metrics
}

func NewFiler(storage metric.Storage, fileName string) (*Filer, error) {

	if fileName == "" {
		return nil, fmt.Errorf("%w: fileName = %s", ErrEmptyFilePath, fileName)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fileName = filepath.Join(cwd, fileName)

	return &Filer{
		fileName: fileName,
		storage:  storage,
	}, nil
}

/*
Load - получаем данные метрик из хранилища снимков, и загружаем их в основное хранилище

TODO чтобы такие "толстые" методы можно было покрыть тестами (с хорошим уровнем покрытия), надо из разбить на мелкие методы.
  - Тогда можно будет протестировать логику метода через моки/стабы
*/
func (filer *Filer) Load(ctx context.Context) (err error) {
	defer func() {
		errClose := filer.closeFile()
		err = errors.Join(err, errClose)
	}()

	err = filer.openFile()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
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

	metrics := make([]models.Metrics, 0, 30)
	for decoder.More() {
		modelMetric := models.Metrics{}
		err = decoder.Decode(&modelMetric)

		if err != nil {
			return err
		}
	}

	if len(metrics) > 0 {
		err = filer.storage.UpdateBatch(ctx, metrics)
		if err != nil {
			return err
		}
	}

	// Читаем закрывающую скобку массива
	if _, err = decoder.Token(); err != nil {
		return err
	}

	return nil
}
func (filer *Filer) Save(ctx context.Context) error {
	err := filer.readFromStorage(ctx)
	if err != nil {
		return err
	}

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
func (filer *Filer) readFromStorage(ctx context.Context) error {
	gaugeMetrics, err := filer.storage.MetricList(ctx, models.Gauge)
	if err != nil {
		return err
	}

	counterMetrics, err := filer.storage.MetricList(ctx, models.Counter)
	if err != nil {
		return err
	}

	if len(gaugeMetrics) == 0 && len(counterMetrics) == 0 {
		return nil
	}

	filer.metrics = make([]models.Metrics, 0, len(gaugeMetrics)+len(counterMetrics))

	if len(gaugeMetrics) > 0 {
		for _, modelMetric := range gaugeMetrics {
			filer.metrics = append(filer.metrics, modelMetric)
		}
	}

	if len(counterMetrics) > 0 {
		for _, modelMetric := range counterMetrics {
			filer.metrics = append(filer.metrics, modelMetric)
		}
	}
	return nil
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
