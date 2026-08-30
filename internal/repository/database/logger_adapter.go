package database

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/jackc/pgx/v5/tracelog"
)

// PgxLoggerAdapter адаптирует logger.Logger к tracelog.TraceLog
type PgxLoggerAdapter struct {
	log      logger.Logger
	logLevel tracelog.LogLevel
}

// NewPgxLoggerAdapter создаёт адаптер
func NewPgxLoggerAdapter(log logger.Logger, logLevel tracelog.LogLevel) *PgxLoggerAdapter {
	return &PgxLoggerAdapter{
		log:      log,
		logLevel: logLevel,
	}
}

// Log реализует интерфейс pgx.Logger
func (pg *PgxLoggerAdapter) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	_ = ctx

	// Если есть SQL и args — подставляем значения для отладки
	if pg.logLevel == tracelog.LogLevelDebug {
		data = sqlFormatted(data)
	}

	// Преобразуем tracelog.LogLevel в уровень logger.Logger
	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		pg.log.Debug(msg, "data", data)
	case tracelog.LogLevelInfo:
		pg.log.Info(msg, "data", data)
	case tracelog.LogLevelWarn:
		pg.log.Warn(msg, "data", data)
	case tracelog.LogLevelError:
		pg.log.Error(msg, "data", data)
	default:
		pg.log.Info(msg, "data", data)
	}
}

func sqlFormatted(data map[string]interface{}) map[string]interface{} {
	if sql, ok := data["sql"].(string); ok {
		if args, ok := data["args"].([]any); ok && len(args) > 0 {
			// Создаём копию data с форматированным SQL
			newData := make(map[string]interface{})
			for k, v := range data {
				newData[k] = v
			}
			newData["sql_formatted"] = formatSQL(sql, args)
			data = newData
		}
	}
	return data
}

// formatSQL заменяет плейсхолдеры на реальные значения (ТОЛЬКО ДЛЯ ЛОГОВ!)
func formatSQL(sql string, args []any) string {
	result := sql
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)

		// Проверяем на nil (включая указатели)
		if isNil(arg) {
			result = strings.Replace(result, placeholder, "NULL", 1)
			continue
		}

		switch v := arg.(type) {
		case string:
			result = strings.Replace(result, placeholder, fmt.Sprintf("'%s'", v), 1)
		case int, int64, int32, float64, float32:
			result = strings.Replace(result, placeholder, fmt.Sprintf("%v", v), 1)
		case []byte:
			result = strings.Replace(result, placeholder, fmt.Sprintf("'%s'", string(v)), 1)
		default:
			result = strings.Replace(result, placeholder, fmt.Sprintf("%v", v), 1)
		}
	}
	return result
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		return rv.IsNil()
	default:
		return false
	}
}
