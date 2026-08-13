package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

// responseData - структура для хранения сведений о запросе
type requestData struct {
	URI     string              `json:"uri"`
	Method  string              `json:"method"`
	Body    string              `json:"body"`
	Headers map[string][]string `json:"headers"`
}

// responseData - структура для хранения сведений об ответе
type responseData struct {
	Status   int           `json:"status"`
	Size     int           `json:"size"`
	Duration time.Duration `json:"duration"`
}

// WithLogging - добавляет дополнительный код для регистрации сведений о запросе и возвращает новый http.Handler.
func WithLogging(logger logger.Logger, h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Читаем тело для логирования
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = io.ReadAll(r.Body)
			_ = r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		respData := &responseData{
			Status: http.StatusOK,
			Size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   respData,
		}

		// внедряем реализацию http.ResponseWriter
		// таким образом будут вызываться методы: loggingResponseWriter.Write(), loggingResponseWriter.WriteHeader().
		h.ServeHTTP(&lw, r)

		respData.Duration = time.Since(start)

		reqData := requestData{
			URI:     r.RequestURI,
			Method:  r.Method,
			Body:    string(bodyBytes),
			Headers: r.Header,
		}

		logger.Info(reqData.Method, "request", reqData, "response", respData)
	}
	return http.HandlerFunc(logFn)
}

// loggingResponseWriter - Добавляем реализацию http.ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter // Встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.Size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.Status = statusCode // захватываем код статуса
}
