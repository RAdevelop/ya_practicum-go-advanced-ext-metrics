package logger

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Logger - интерфейс для логирования
//
//go:generate mockery
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func New() Logger {
	return &LogMe{
		log: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}
}

type LogMe struct {
	log *slog.Logger
}

func (l *LogMe) Debug(msg string, args ...any) {
	l.log.Debug(msg, l.withFileLine(args...)...)
}
func (l *LogMe) Info(msg string, args ...any) {
	l.log.Info(msg, l.withFileLine(args...)...)
}
func (l *LogMe) Warn(msg string, args ...any) {
	l.log.Warn(msg, l.withFileLine(args...)...)
}
func (l *LogMe) Error(msg string, args ...any) {
	l.log.Error(msg, l.withFileLine(args...)...)
}

// withFileLine - автоматически находит вызывающий код, игнорируя внутренние пакеты
func (l *LogMe) withFileLine(args ...any) []any {
	// Пропускаем уровни, пока не найдём файл вне пакета logger
	for skip := 2; skip < 10; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		// Если файл не содержит "logger" и не является внутренним для Go (net/http, testing и т.д.)
		if !strings.Contains(file, "/logger/") && !strings.Contains(file, "logger.go") &&
			!strings.Contains(file, "/net/http/") && !strings.Contains(file, "/testing/") {
			caller := file + ":" + strconv.Itoa(line)
			return append(args, slog.String("caller", caller))
		}
	}
	// fallback
	return args
}
