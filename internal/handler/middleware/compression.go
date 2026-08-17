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
		isClientAcceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		// Проверяем, поддерживает ли клиент gzip
		if !isClientAcceptGzip {
			next.ServeHTTP(w, r)
			return
		}

		gw := newGzipResponseWriter(isClientAcceptGzip, w, logger)
		defer func() {
			err := gw.Close()
			if err != nil {
				logger.Error("GzipResponseWriter", "err", err)
			}
		}()

		next.ServeHTTP(gw, r)
	})
}

// Decompression - распаковываем данные, если клиент прислал их запакованные в gzip
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
	"text/html":        true,
	"application/json": true,
}

// shouldDecompress - проверяем, нужно ли будет распаковать данные
func shouldDecompress(r *http.Request) bool {
	return r.Body != nil && r.Header.Get("Content-Encoding") == "gzip"
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz                 *gzip.Writer
	logger             logger.Logger
	isClientAcceptGzip bool
	isHeaderWritten    bool
}

func newGzipResponseWriter(isClientAcceptGzip bool, w http.ResponseWriter, logger logger.Logger) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter:     w,
		gz:                 nil,
		logger:             logger,
		isClientAcceptGzip: isClientAcceptGzip,
		isHeaderWritten:    false,
	}
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if gw.shouldCompress() {
		gw.initGzipWriter()
		gw.setHeaderContentEncoding()

		//сжимаем
		return gw.gz.Write(b)
	}

	return gw.ResponseWriter.Write(b)
}

// WriteHeader переопределяем, чтобы установить Content-Length после сжатия
func (gw *gzipResponseWriter) WriteHeader(statusCode int) {
	gw.setHeaderContentEncoding()
	gw.ResponseWriter.WriteHeader(statusCode)
}

func (gw *gzipResponseWriter) Close() error {
	if gw.gz != nil {
		// Сначала сбрасываем данные
		if err := gw.gz.Flush(); err != nil {
			return err
		}

		// Затем закрываем
		if err := gw.gz.Close(); err != nil {
			return err
		}

		gzipPool.Put(gw.gz)
		gw.gz = nil
		return nil
	}
	return nil
}

// shouldCompress - проверяем, будем ли сжимать данные
func (gw *gzipResponseWriter) shouldCompress() bool {
	if !gw.isClientAcceptGzip {
		return false
	}

	contentType := strings.Split(gw.ResponseWriter.Header().Get("Content-Type"), ";")[0]
	needCompress, exists := contentTypesForCompression[contentType]

	if !exists {
		return false
	}

	return needCompress
}

func (gw *gzipResponseWriter) setHeaderContentEncoding() {
	if gw.shouldCompress() && !gw.isHeaderWritten {
		gw.isHeaderWritten = true
		// Сообщаем клиенту, что данные запакованы
		gw.ResponseWriter.Header().Set("Content-Encoding", "gzip")
		gw.ResponseWriter.Header().Set("Vary", "Accept-Encoding")
	}
}

func (gw *gzipResponseWriter) initGzipWriter() {
	// инициализируем его только когда точно знаем, что надо будет сжимать ответ
	if gw.gz == nil {
		// Получаем/Создаем Writer из pool
		gw.gz = gzipPool.Get().(*gzip.Writer)

		/*
			gz.Reset():
			- Перенаправляет вывод на новый io.Writer() (в данном случае http.ResponseWriter())
			- Сбрасывает внутреннее состояние (не создавая новый объект)
			- Готовит gzip.Writer() к использованию с новым потоком данных
		*/
		gw.gz.Reset(gw.ResponseWriter)
	}
}
