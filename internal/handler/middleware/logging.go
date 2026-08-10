package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

// responseData - структура для хранения сведений о запросе
type requestData struct {
	Uri    string `json:"uri"`
	Method string `json:"method"`
}

// responseData - структура для хранения сведений об ответе
type responseData struct {
	Status   int           `json:"status"`
	Size     int           `json:"size"`
	Duration time.Duration `json:"duration"`
}

// WithLogging - добавляет дополнительный код для регистрации сведений о запросе и возвращает новый http.Handler.
func WithLogging(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

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
			Uri:    r.RequestURI,
			Method: r.Method,
		}
		logging(respData, reqData)
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

func logging(respData *responseData, reqData requestData) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info(reqData.Method, slog.Any("request", reqData), slog.Any("response", respData))
}
