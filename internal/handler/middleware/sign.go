package middleware

import (
	"bytes"
	"io"
	"net/http"
	"sync"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/appcontext"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/headers"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/sign"
)

// SignCheck - проверяет, есть ли заголовок "HashSHA256" со значением, если есть, то проверяет подпись
func SignCheck(appContext *appcontext.AppContext, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if appContext.Config.SignKey() == "" {
			next.ServeHTTP(w, r)
			return
		}

		receivedHash := r.Header.Get(headers.HashHeader)
		if receivedHash == "" {
			http.Error(w, "missing sign header", http.StatusBadRequest)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			appContext.Logger.Error("cannot read body", "error", err)
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}

		if !sign.SHA256Verify(body, appContext.Config.SignKey(), receivedHash) {
			http.Error(w, "cannot verify sign", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(body))
		next.ServeHTTP(w, r)
	})
}

// SignAdd — middleware, который подписывает тело ответа HMAC-SHA256, если клиент прислал заголовок HashSHA256.
func SignAdd(appContext *appcontext.AppContext, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if appContext.Config.SignKey() == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Оборачиваем ResponseWriter в буфер
		sw := newSignResponseWriter(w, appContext)
		defer func() {
			if err := sw.Close(); err != nil {
				appContext.Logger.Error("SignResponseWriter close", "err", err)
			}
		}()

		next.ServeHTTP(sw, r)
	})
}

/*
signPool - для экономии выделяемой памяти на каждом http запросе при использовании данной middleware
  - Хранит временные объекты для переиспользования
  - При Get() — возвращает объект из пула (или создает новый через New)
  - При Put() — возвращает объект в пул для будущего использования
  - GC автоматически очищает пул при необходимости
*/
var signPool = sync.Pool{
	New: func() any {
		return &bytes.Buffer{}
	},
}

type signResponseWriter struct {
	http.ResponseWriter
	appContext      *appcontext.AppContext
	buf             *bytes.Buffer
	statusCode      int
	isHeaderWritten bool
}

func newSignResponseWriter(w http.ResponseWriter, appContext *appcontext.AppContext) *signResponseWriter {
	return &signResponseWriter{
		ResponseWriter: w,
		appContext:     appContext,
		buf:            signPool.Get().(*bytes.Buffer),
		statusCode:     http.StatusOK,
	}
}

// Write копит тело в буфере, не отправляя его клиенту
func (s *signResponseWriter) Write(b []byte) (int, error) {
	return s.buf.Write(b)
}

// WriteHeader запоминает статус, но не отправляет его сразу
func (s *signResponseWriter) WriteHeader(statusCode int) {
	if s.isHeaderWritten {
		return
	}
	s.isHeaderWritten = true
	s.statusCode = statusCode
}

// Close считает подпись, записывает заголовок и отправляет тело
func (s *signResponseWriter) Close() error {
	body := s.buf.Bytes()

	// Считаем подпись от всего тела
	hash := sign.SHA256(body, s.appContext.Config.SignKey())

	// Устанавливаем заголовок (до WriteHeader!)
	s.ResponseWriter.Header().Set(headers.HashHeader, hash)

	// Отправляем статус
	s.ResponseWriter.WriteHeader(s.statusCode)

	// Отправляем тело
	_, err := s.ResponseWriter.Write(body)

	// Возвращаем буфер в пул
	s.buf.Reset()
	signPool.Put(s.buf)
	return err
}
