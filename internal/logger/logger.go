package logger

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
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

// withFileLine - добавим ко всем логам путь к файлу и номеру строки, где вызвали метод логирования
func (l *LogMe) withFileLine(args ...any) []any {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "unknown_file"
		line = 0
	}

	caller := file + ":" + strconv.Itoa(line)

	return append(args, slog.String("caller", caller))
}
