package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

/*
gzipPool - для экономии выделяемой памяти на каждом http запросе при использовании данной middleware
  - Хранит временные объекты для переиспользования
  - При Get() — возвращает объект из пула (или создает новый через New)
  - При Put() — возвращает объект в пул для будущего использования
  - GC автоматически очищает пул при необходимости
*/
var gzipPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return gz
	},
}

// Compression - если клиент поддерживает прием gzip данных, то сжимаем их перед ответом клиенту
func Compression(logger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Проверяем, поддерживает ли клиент gzip
		if !shouldCompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		// Получаем Writer из pool
		gz := gzipPool.Get().(*gzip.Writer)
		defer func() {
			err := gz.Close()
			if err != nil {
				logger.Error("error closing gzip writer", "error", err)
			}
			gzipPool.Put(gz) // Возвращаем в pool
		}()

		/*
			gz.Reset():
			- Перенаправляет вывод на новый io.Writer() (в данном случае http.ResponseWriter())
			- Сбрасывает внутреннее состояние (не создавая новый объект)
			- Готовит gzip.Writer() к использованию с новым потоком данных
		*/
		gz.Reset(w)
		// Сообщаем клиенту, что данные запакованы
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		gw := &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
		}

		next.ServeHTTP(gw, r)
	})
}

// Decompression - распаковываем данные, если клиент прислал их запакованные в gizp
func Decompression(logger logger.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldDecompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			logger.Error("error create gzip reader", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		defer func() {
			err := gz.Close()
			if err != nil {
				logger.Error("error closing gzip reader", "error", err)
			}
		}()

		// Читаем распакованные данные
		decompressedBody, err := io.ReadAll(gz)
		if err != nil {
			logger.Error("failed to read gzip data", "error", err)
			http.Error(w, "Failed to read gzip data", http.StatusBadRequest)
			return
		}

		// Закрываем оригинальное тело
		err = r.Body.Close()
		if err != nil {
			logger.Error("failed to close body", "error", err)
		}

		// Подменяем тело на распакованные данные
		r.Body = io.NopCloser(bytes.NewReader(decompressedBody))
		r.ContentLength = int64(len(decompressedBody))

		// Удаляем заголовок Content-Encoding (тело уже распаковано)
		r.Header.Del("Content-Encoding")

		// Передаем управление обработчику
		next.ServeHTTP(w, r)
	})
}

// contentTypesForCompression - список типов контента, для которых [не]надо будет сжимать данные
var contentTypesForCompression = map[string]bool{
	"text/html":                true,
	"text/html; charset=utf-8": true,
	"application/json":         true,
	"":                         true, //если клиент не прислал тип контента
}

// shouldCompress - проверяем, поддерживает ли клиент gzip
func shouldCompress(r *http.Request) bool {

	contentType := r.Header.Get("Content-Type")
	needCompress, exists := contentTypesForCompression[contentType]

	if !exists {
		return false
	}

	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && needCompress
}

// shouldCompress - проверяем, поддерживает ли клиент gzip
func shouldDecompress(r *http.Request) bool {
	return r.Body != nil && r.Header.Get("Content-Encoding") == "gzip"
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	return gw.gz.Write(b)
}

// WriteHeader переопределяем, чтобы установить Content-Length после сжатия
func (gw *gzipResponseWriter) WriteHeader(statusCode int) {
	gw.ResponseWriter.WriteHeader(statusCode)
}
