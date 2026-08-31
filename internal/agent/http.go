package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"net/http"
	"sync"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/retryer"
	"github.com/go-resty/resty/v2"
)

// HttpAgent - http клиент для отправки метрик на сервер
type HttpAgent struct {
	client *resty.Client
}

func New(client *resty.Client) *HttpAgent {
	return &HttpAgent{
		client: client,
	}
}

func (a HttpAgent) Update(ctx context.Context, metric models.Metrics) (*http.Response, error) {
	url := "/update"
	return a.sendPostJsonRetryLinear(ctx, url, metric)

}

func (a HttpAgent) Updates(ctx context.Context, metrics []models.Metrics) (*http.Response, error) {
	url := "/updates"

	return a.sendPostJsonRetryLinear(ctx, url, metrics)
}

func (a HttpAgent) sendPostJsonRetryLinear(ctx context.Context, url string, v any) (*http.Response, error) {
	return retryer.RetryLinear(ctx, func(ctx context.Context) (*http.Response, error) {
		return a.sendPostJson(ctx, url, v)
	}, 2, new(3))
}

func (a HttpAgent) sendPostJson(ctx context.Context, url string, v any) (*http.Response, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	body, err = compress(body)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetDoNotParseResponse(true).
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(body).
		Post(url)

	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
}

// gzipPool - Пул для gzip.Writer() (сжимает данные)
var gzipPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(nil, gzip.DefaultCompression)
		return gz
	},
}

// bufferPool - Пул для bytes.Buffer() (хранит сжатые данные)
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// Compress - сжимает данные с использованием пулов
func compress(data []byte) ([]byte, error) {
	// Берем буфер из пула
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf) // Возвращаем в пул после использования

	// Берем gzip.Writer() из пула
	gz := gzipPool.Get().(*gzip.Writer)
	gz.Reset(buf) // Направляем вывод в буфер
	defer func() {
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
