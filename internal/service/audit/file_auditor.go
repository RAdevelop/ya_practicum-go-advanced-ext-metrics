package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileAuditor — дописывает каждое событие в файл на новой строке.
type FileAuditor struct {
	f  *os.File
	mu sync.Mutex
}

func NewFileAuditor(path string) (*FileAuditor, error) {

	// Получаем директорию файла
	dir := filepath.Dir(path)

	// Создаем директорию, если её нет
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open audit file: %w", err)
	}

	return &FileAuditor{f: f}, nil
}

// Save - сохраняет данные о событии в источник хранения данных аудитора
func (a *FileAuditor) Save(_ context.Context, event Event) error {

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	data = append(data, '\n')

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.f == nil {
		return errors.New("audit file is closed")
	}

	if _, err = a.f.Write(data); err != nil {
		return fmt.Errorf("write audit file: %w", err)
	}

	return nil
}

// Close - закрывает файловый дескриптор.
// Вызывать один раз при штатном завершении сервиса (в Shutdown-пути приложения).
func (a *FileAuditor) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.f == nil {
		return nil
	}
	defer func() { a.f = nil }()
	return a.f.Close()
}
