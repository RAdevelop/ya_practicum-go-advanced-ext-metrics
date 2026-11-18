package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileAuditor — дописывает каждое событие в файл на новой строке.
type FileAuditor struct {
	path string
	mu   sync.Mutex
}

func NewFileAuditor(path string) (*FileAuditor, error) {

	// Получаем директорию файла
	dir := filepath.Dir(path)

	// Создаем директорию, если её нет
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &FileAuditor{path: path}, nil
}

func (s *FileAuditor) Save(ctx context.Context, event Event) error {

	_ = ctx

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}
	data = append(data, '\n')

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write audit file: %w", err)
	}

	return nil
}
