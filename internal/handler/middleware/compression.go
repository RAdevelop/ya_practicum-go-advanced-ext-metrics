package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/appcontext"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

// minSizeToCompress — минимальный размер ответа, при котором имеет смысл сжимать.
// Меньшие ответы сжимать невыгодно: overhead gzip больше выигрыша.
const minSizeToCompress = 3_000

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
func Compression(appContext *appcontext.AppContext, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isClientAcceptGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

		// Проверяем, поддерживает ли клиент gzip
		if !isClientAcceptGzip {
			next.ServeHTTP(w, r)
			return
		}

		gw := newGzipResponseWriter(isClientAcceptGzip, w, appContext.Logger)
		defer func() {
			err := gw.Close()
			if err != nil {
				appContext.Logger.Error("GzipResponseWriter", "err", err)
			}
		}()

		next.ServeHTTP(gw, r)
	})
}

// Decompression - распаковываем данные, если клиент прислал их запакованные в gzip
func Decompression(appContext *appcontext.AppContext, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldDecompress(r) {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			appContext.Logger.Error("error create gzip reader", "error", err)
			next.ServeHTTP(w, r)
			return
		}
		defer func() {
			err := gz.Close()
			if err != nil {
				appContext.Logger.Error("error closing gzip reader", "error", err)
			}
		}()

		// Читаем распакованные данные
		decompressedBody, err := io.ReadAll(gz)
		if err != nil {
			appContext.Logger.Error("failed to read gzip data", "error", err)
			http.Error(w, "Failed to read gzip data", http.StatusBadRequest)
			return
		}

		// Закрываем оригинальное тело
		err = r.Body.Close()
		if err != nil {
			appContext.Logger.Error("failed to close body", "error", err)
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

	// isHeaderWritten — заголовки уже отправлены в реальный ResponseWriter.
	// Выставляется во всех путях, где происходит первый WriteHeader/Write в underlying writer:
	// и при сжатии (перед этим добавляется Content-Encoding), и при прямом проксировании ответа.
	isHeaderWritten bool

	buf           []byte // буфер для накопления до minSizeToCompress
	decided       bool   // решили ли: сжимать или нет
	pendingStatus int    // отложенный WriteHeader, пока не решили
}

func newGzipResponseWriter(isClientAcceptGzip bool, w http.ResponseWriter, logger logger.Logger) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter:     w,
		logger:             logger,
		isClientAcceptGzip: isClientAcceptGzip,
	}
}

// writeThroughHeader отправляет статус в реальный ResponseWriter ровно один раз.
// Все пути (сжимаемый и несжимаемый) идут через него, чтобы исключить двойной WriteHeader.
func (gw *gzipResponseWriter) writeThroughHeader(code int) {
	if gw.isHeaderWritten {
		return
	}
	gw.isHeaderWritten = true
	gw.ResponseWriter.WriteHeader(code)
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if !gw.shouldCompress() {
		// Несжимаемый ответ: если статус был отложен — отправляем его перед записью тела,
		// иначе net/http неявно отправит 200 и ошибка клиента превратится в 200.
		if gw.pendingStatus != 0 {
			st := gw.pendingStatus
			gw.pendingStatus = 0
			gw.writeThroughHeader(st)
		}
		// Запись в underlying writer неявно отправляет заголовки
		gw.isHeaderWritten = true
		return gw.ResponseWriter.Write(b)
	}

	if !gw.decided {
		gw.buf = append(gw.buf, b...)
		if len(gw.buf) < minSizeToCompress {
			return len(b), nil
		}

		// накопили достаточно — сжимаем
		gw.decided = true
		gw.initGzipWriter()
		gw.setHeaderContentEncoding()
		if gw.pendingStatus != 0 {
			st := gw.pendingStatus
			gw.pendingStatus = 0
			gw.writeThroughHeader(st)
		}
		_, err := gw.gz.Write(gw.buf)
		gw.buf = nil
		return len(b), err
	}

	return gw.gz.Write(b)
}

// WriteHeader переопределяем, чтобы установить Content-Encoding до отправки заголовка.
// Если ещё не решили, сжимать или нет — откладываем.
func (gw *gzipResponseWriter) WriteHeader(statusCode int) {
	if gw.isHeaderWritten {
		return // защита от повторного вызова ("superfluous response.WriteHeader call")
	}
	if gw.decided {
		gw.setHeaderContentEncoding()
		gw.writeThroughHeader(statusCode)
		return
	}
	gw.pendingStatus = statusCode
}

func (gw *gzipResponseWriter) Close() error {
	if gw.decided && gw.gz != nil {
		// Сначала сбрасываем данные
		if err := gw.gz.Flush(); err != nil {
			return err
		}

		// Затем закрываем
		if err := gw.gz.Close(); err != nil {
			return err
		}

		gw.gz.Reset(io.Discard) // сброс ссылки на ResponseWriter
		gzipPool.Put(gw.gz)
		gw.gz = nil
		return nil
	}

	// не сжимали — если статус ещё не отправлен, отправляем отложенный
	if gw.pendingStatus != 0 {
		st := gw.pendingStatus
		gw.pendingStatus = 0
		gw.writeThroughHeader(st)
	}
	// пишем сырые данные, если они ещё не ушли
	if len(gw.buf) > 0 {
		gw.isHeaderWritten = true // запись отправит заголовки неявно
		_, err := gw.ResponseWriter.Write(gw.buf)
		gw.buf = nil
		return err
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
