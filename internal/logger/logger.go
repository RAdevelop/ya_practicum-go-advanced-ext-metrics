package logger

import (
	"log/slog"
	"os"
)

// Logger - интерфейс для логирования
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func New() Logger {
	return &LogMe{
		log: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

type LogMe struct {
	log *slog.Logger
}

func (l *LogMe) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}
func (l *LogMe) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}
func (l *LogMe) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}
func (l *LogMe) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

// LogMeTest - заглушка для логов во время выполнения тестов
type LogMeTest struct{}

func NewTest() Logger {
	return &LogMeTest{}
}

func (l *LogMeTest) Debug(string, ...any) {}
func (l *LogMeTest) Info(string, ...any)  {}
func (l *LogMeTest) Warn(string, ...any)  {}
func (l *LogMeTest) Error(string, ...any) {}
