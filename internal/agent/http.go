package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/headers"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/retryer"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/sign"
	"github.com/go-resty/resty/v2"
)

// HTTPAgent — HTTP-клиент агента для отправки метрик на сервер.
//
// Использует resty, сжимает тело запроса в gzip
// и повторяет отправку при временных ошибках.
//
// Внутренние поля:
//
//	client — настроенный resty.Client (base URL, таймауты, ретраи на уровне транспорта);
//	config — провайдер конфигурации агента: адрес сервера и ключ подписи
type HTTPAgent struct {
	client *resty.Client
	config agent.ConfigProvider
}

// New — конструктор агента. Принимает настроенный resty.Client
// и провайдер конфигурации (адрес сервера, ключ подписи).
func New(client *resty.Client, config agent.ConfigProvider) *HTTPAgent {
	return &HTTPAgent{
		client: client,
		config: config,
	}
}

// Update — отправляет одну метрику на сервер: POST /update (JSON, gzip, ретраи).
func (a HTTPAgent) Update(ctx context.Context, metric models.Metrics) (*http.Response, error) {
	url := "/update"
	return a.sendPostJSONRetryLinear(ctx, url, metric)

}

// Updates — отправляет пакет метрик на сервер: POST /updates/ (JSON, gzip, ретраи).
func (a HTTPAgent) Updates(ctx context.Context, metrics []models.Metrics) (*http.Response, error) {
	url := "/updates/"

	return a.sendPostJSONRetryLinear(ctx, url, metrics)
}

func (a HTTPAgent) sendPostJSONRetryLinear(ctx context.Context, url string, v any) (*http.Response, error) {
	return retryer.RetryLinear(ctx, func(ctx context.Context) (*http.Response, error) {
		return a.sendPostJSON(ctx, url, v)
	}, 2, new(3))
}

func (a HTTPAgent) sendPostJSON(ctx context.Context, url string, v any) (*http.Response, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	body, err = compress(body)
	if err != nil {
		return nil, err
	}

	req := a.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetDoNotParseResponse(true).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(body)

	if a.config.SignKey() != "" {
		req.SetHeader(headers.HashHeader, sign.SHA256(body, a.config.SignKey()))
	}

	resp, err := req.Post(url)

	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

// gzipPool — пул gzip.Writer для переиспользования между запросами
// и снижения количества аллокаций.
var gzipPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return gz
	},
}

// bufferPool — пул bytes.Buffer для накопления сжатых данных.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// compress — сжимает данные gzip с использованием пулов writer'ов и буферов.
func compress(data []byte) ([]byte, error) {
	// Берем буфер из пула
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf) // Возвращаем в пул после использования

	// Берем gzip.Writer() из пула
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(buf) // Направляем вывод в буфер
	defer func() {
		// Сбрасываем ссылку на внешний writer (io.Discard — безопасная заглушка).
		// Внутренние буферы gzip сохраняются для переиспользования из пула.
		gz.Reset(io.Discard)
		gzipPool.Put(gz) // Возвращаем в пул
	}()

	// Сжимаем данные
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	// Копируем результат (обязательно, т.к. буфер будет переиспользован)
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())
	return result, nil
}
