package router

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type params struct {
	metricType  string
	metricName  string
	metricValue string
	url         string
}
type want struct {
	contentType string
	statusCode  int
	metricValue any
}
type given struct {
	contentType string
	method      string
	reqParams   params
}

func setupMockLogger(t *testing.T) *logger.MockLogger {
	logMe := logger.NewMockLogger(t)

	//не знаю как лучше сделать возможное переменное количество параметров для вызова таких методов... :(
	logMe.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Warn(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Debug(mock.Anything, mock.Anything, mock.Anything).Maybe()

	return logMe
}

func setupMockConfigProvider(t *testing.T) *server.MockConfigProvider {
	mock := server.NewMockConfigProvider(t)

	mock.EXPECT().FileStoragePath().Maybe().Return("mock.file")
	mock.EXPECT().Address().Maybe().Return("localhost:8080")
	mock.EXPECT().StoreInterval().Maybe().Return(nil)
	mock.EXPECT().Restore().Maybe().Return(nil)

	return mock
}

func TestMetric_UpdateWithTextPlain(t *testing.T) {

	tests := []struct {
		name  string
		given given
		want  want
	}{
		//metric Counter
		{
			name: "counter metric update with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("counter", "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "counter metric update with StatusMethodNotAllowed",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForUpdate("counter", "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
		{
			name: "counter metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "application/json", //StatusUnsupportedMediaType
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("counter", "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("unknownMetricType" /*причина StatusBadRequest*/, "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("123counterBadName" /*причина StatusBadRequest*/, "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusNotFound",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("counter", "123someMetricBadName" /*причина StatusNotFound*/, "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("counter", "someMetric", "s527" /*причина StatusBadRequest*/),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},

		//metric Gauge
		{
			name: "gauge metric update with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("gauge", "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "gauge metric update with StatusMethodNotAllowed",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForUpdate("gauge", "someMetric", "728.4110804980572"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusMethodNotAllowed,
			},
		},
		{
			name: "gauge metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "application/json", //StatusUnsupportedMediaType
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("gauge", "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "gauge metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("unknownMetricType" /*причина StatusBadRequest*/, "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "gauge metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("123gaugeBadName" /*причина StatusBadRequest*/, "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "gauge metric update with StatusNotFound",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("gauge", "123someMetricBadName" /*причина StatusNotFound*/, "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name: "gauge metric update with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams:   metricBuildParamsForUpdate("gauge", "someMetric", "s527.123" /*причина StatusBadRequest*/),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
	}

	var err error

	loggerTest := setupMockLogger(t)
	mockConfigProvider := setupMockConfigProvider(t)
	memStorage := memory.NewStorage()
	metricService := metric.NewService(memStorage)
	metricSnapshot := snapshot.NewMockAble(t)
	var metricManager = service.NewManager(metricService, metricSnapshot)
	h := handler.New(metricManager, loggerTest, mockConfigProvider)
	r := New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()

	// Создаем resty-клиент с тестовым URL
	client := resty.New()
	client.SetBaseURL(mockServer.URL)

	var result *resty.Response

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := client.R().
				SetHeader("Content-Type", tt.given.contentType).
				SetDoNotParseResponse(true)

			switch tt.given.method {
			case http.MethodPost:
				result, err = req.Post(tt.given.reqParams.url)
			case http.MethodGet:
				result, err = req.Get(tt.given.reqParams.url)
			}

			assert.Equalf(t, tt.want.statusCode, result.StatusCode(), "tt.given.reqParams: %v", tt.given.reqParams)
			// io.Discard выступает в качестве приёмника ненужных данных
			_, err = io.Copy(io.Discard, result.RawResponse.Body)
			assert.NoError(t, err)
			assert.NoError(t, result.RawResponse.Body.Close())
		})
	}
}

func TestMetric_UpdateWithJson(t *testing.T) {

	type given struct {
		body string
	}
	type want struct {
		body        string
		contentType string
		statusCode  int
	}
	tests := []struct {
		name  string
		given given
		want  want
	}{
		//gauge
		{
			name: "json metric update gauge with StatusOK",
			given: given{
				body: `{"id":"LastGC","type":"gauge","value":1744184459}`,
			},
			want: want{
				body:        `{"id":"LastGC","type":"gauge","value":1744184459}`,
				contentType: "application/json",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "json metric update gauge with StatusBadRequest",
			given: given{
				body: `{"id": "LastGC","type": "gauge","value": }`,
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "json metric update gauge with StatusBadRequest",
			given: given{
				body: `{"id": "LastGC","type": "badMetricNameGauge","value": 0.00000001}`,
			},
			want: want{
				body: `Metric type "badMetricNameGauge" is not supported.
Use one of the supported metric types: [counter gauge]`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "json metric update gauge with StatusBadRequest",
			given: given{
				body: `{"id": "LastGC","type": "gauge","value": "0.00000001"}`, //value не надо оборачивать кавычками
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "json metric update gauge with StatusBadRequest",
			given: given{
				body: `{"id": "LastGC","type": "gauge","value": ""}`, //value не надо оборачивать кавычками
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		// counter
		{
			name: "counter metric update with StatusOK",
			given: given{
				body: `{"id":"someMetric","type":"counter","delta":527}`,
			},
			want: want{
				body:        `{"id":"someMetric","type":"counter","delta":527}`,
				contentType: "application/json",
				statusCode:  http.StatusOK,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				body: `{"id": "someMetric","type": "counter","delta": ""}`,
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				body: `{"id": "someMetric","type": "counter","delta": }`,
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				body: `{"id": "someMetric","type": "counter","delta"}`, //не валидный json
			},
			want: want{
				body:        `Can't parse request body`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusBadRequest",
			given: given{
				body: `{"id": "someMetric","type": "badMetricType","delta":123}`, //тип метрики не: counter, gauge
			},
			want: want{
				body: `Metric type "badMetricType" is not supported.
Use one of the supported metric types: [counter gauge]`,
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
			},
		},
	}

	var err error
	loggerTest := setupMockLogger(t)
	mockConfigProvider := setupMockConfigProvider(t)
	memStorage := memory.NewStorage()
	metricService := metric.NewService(memStorage)
	metricSnapshot := snapshot.NewMockAble(t)
	var metricManager = service.NewManager(metricService, metricSnapshot)
	h := handler.New(metricManager, loggerTest, mockConfigProvider)
	r := New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()

	// Создаем resty-клиент с тестовым URL
	client := resty.New()
	client.SetBaseURL(mockServer.URL)

	var result *resty.Response

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetDoNotParseResponse(true).
				SetBody(tt.given.body)

			result, err = req.Post("/update")

			assert.NoErrorf(t, err, "tt.given: %v", tt.given)
			assert.Equalf(t, tt.want.statusCode, result.StatusCode(), "tt.given: %v", tt.given)
			assert.Equalf(t, tt.want.contentType, result.Header().Get("Content-Type"), "tt.given: %v", tt.given)
			body, err := io.ReadAll(result.RawResponse.Body)
			assert.NoErrorf(t, err, "tt.given: %v", tt.given)
			assert.Equalf(t, tt.want.body, strings.TrimSpace(string(body)), "tt.given: %v", tt.given)
		})
	}
}

func TestMetric_GetWithTextPlain(t *testing.T) {

	var metricStorage = memory.NewStorage()
	var metricService = metric.NewService(metricStorage)

	tests := []struct {
		name  string
		given given
		want  want
	}{
		//metric Counter
		{
			name: "get metric counter size 1 with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForCounterGetValue(metricService, "counter", "someMetric", []int64{527}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				metricValue: "527",
			},
		},
		{
			name: "get metric counter size 2 with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForCounterGetValue(metricService, "counter", "someMetric2", []int64{1, 2}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				metricValue: "3",
			},
		},
		{
			name: "get metric counter with StatusNotFound",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForCounterGetValue(metricService, "counter", "notFound" /*причина StatusNotFound*/, []int64{} /*причина StatusNotFound*/),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
				metricValue: `Metric value not found by name
`,
			},
		},
		{
			name: "get metric with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForCounterGetValue(metricService, "unknownType" /*причина StatusBadRequest*/, "notFound", []int64{}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				metricValue: `Metric type "unknownType" is not supported.
Use one of the supported metric types: [counter gauge]
`,
			},
		},

		//metric Gauge
		{
			name: "get metric gauge with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForGaugeGetValue(metricService, "gauge", "someMetric", []float64{527.123}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				metricValue: "527.123",
			},
		},
		{
			name: "get metric gauge updated with StatusOK",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForGaugeGetValue(metricService, "gauge", "someMetric2", []float64{527.123, 0.123}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusOK,
				metricValue: "0.123",
			},
		},
		{
			name: "get metric gauge with StatusNotFound",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForGaugeGetValue(metricService, "gauge", "NotFound", []float64{} /*причина StatusBadRequest*/),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
				metricValue: `Metric value not found by name
`,
			},
		},
		{
			name: "get metric with StatusBadRequest",
			given: given{
				contentType: "text/plain",
				method:      http.MethodGet,
				reqParams:   metricBuildParamsForGaugeGetValue(metricService, "unknownType" /*причина StatusBadRequest*/, "NotFound", []float64{}),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
				metricValue: `Metric type "unknownType" is not supported.
Use one of the supported metric types: [counter gauge]
`,
			},
		},
	}
	loggerTest := setupMockLogger(t)
	mockConfigProvider := setupMockConfigProvider(t)
	metricSnapshot := snapshot.NewMockAble(t)
	var metricManager = service.NewManager(metricService, metricSnapshot)
	h := handler.New(metricManager, loggerTest, mockConfigProvider)
	r := New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()

	// Создаем resty-клиент с тестовым URL
	client := resty.New()
	client.SetBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := client.R().
				SetHeader("Content-Type", tt.given.contentType).
				SetDoNotParseResponse(true)

			result, err := req.Get(tt.given.reqParams.url)

			assert.Equalf(t, tt.want.statusCode, result.StatusCode(), "given: %+v", tt.given)

			metricValue, err := io.ReadAll(result.RawResponse.Body)
			assert.Equal(t, tt.want.metricValue, string(metricValue), "given: %+v", tt.given)

			assert.NoError(t, err)
			assert.NoError(t, result.RawResponse.Body.Close())
		})
	}
}

func TestMetric_GetWithJson(t *testing.T) {

	var metricStorage = memory.NewStorage()
	var metricService = metric.NewService(metricStorage)

	type given struct {
		metric *models.Metrics
	}

	type want struct {
		contentType string
		statusCode  int
		body        string
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "get metric gauge json with StatusOK",
			given: given{
				metric: &models.Metrics{
					MType: models.Gauge,
					ID:    "someMetric",
					Value: new(1744184459.0),
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				body:        `{"id":"someMetric","type":"gauge","value":1744184459}`,
			},
		},
		{
			name: "get metric gauge json with StatusOK",
			given: given{
				metric: &models.Metrics{
					MType: models.Counter,
					ID:    "someMetric",
					Delta: new(int64(42)),
				},
			},
			want: want{
				contentType: "application/json",
				statusCode:  http.StatusOK,
				body:        `{"id":"someMetric","type":"counter","delta":42}`,
			},
		},
		{
			name: "get metric gauge json with no body",
			given: given{
				metric: nil,
			},
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				body:        `Can't parse request body`,
			},
		},
	}

	loggerTest := setupMockLogger(t)
	mockConfigProvider := setupMockConfigProvider(t)
	metricSnapshot := snapshot.NewMockAble(t)
	var metricManager = service.NewManager(metricService, metricSnapshot)
	h := handler.New(metricManager, loggerTest, mockConfigProvider)
	r := New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()
	client := resty.New()
	client.SetBaseURL(mockServer.URL)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := client.R().
				SetHeader("Content-Type", "application/json").
				SetDoNotParseResponse(true)

			if tt.given.metric != nil {

				// сами сначала добавляем значения в хранилище данных
				switch tt.given.metric.MType {
				case models.Gauge:
					metricService.GaugeUpdate(tt.given.metric.ID, *tt.given.metric.Value)
				case models.Counter:
					metricService.CounterAdd(tt.given.metric.ID, *tt.given.metric.Delta)
				}

				sendBody, _ := json.Marshal(tt.given.metric)
				req.SetBody(sendBody)
			}

			result, err := req.Post("/value")

			assert.NoError(t, err)
			assert.Equalf(t, tt.want.statusCode, result.StatusCode(), "wrong status code")
			assert.Equalf(t, tt.want.contentType, result.Header().Get("Content-Type"), "wrong Content-Type")

			var body []byte
			if result.RawResponse.Header.Get("Content-Encoding") == "gzip" {
				reader, err := gzip.NewReader(result.RawResponse.Body)
				assert.NoError(t, err)
				defer assert.NoError(t, reader.Close())

				body, err = io.ReadAll(reader)
			} else {
				body, err = io.ReadAll(result.RawResponse.Body)
			}

			assert.NoError(t, err)
			assert.NoError(t, result.RawResponse.Body.Close())
			assert.Equal(t, tt.want.body, strings.TrimSpace(string(body)))

		})
	}
}

func metricBuildParamsForCounterGetValue(metricService *metric.Service, metricType string, metricName string, metricValues []int64) params {

	for _, value := range metricValues {
		metricService.CounterAdd(metricName, value)
	}

	return params{
		metricType: metricType,
		metricName: metricName,
		url:        "/value/" + metricType + "/" + metricName,
	}
}

func metricBuildParamsForGaugeGetValue(metricService *metric.Service, metricType string, metricName string, metricValues []float64) params {
	for _, value := range metricValues {
		metricService.GaugeUpdate(metricName, value)
	}

	return params{
		metricType: metricType,
		metricName: metricName,
		url:        "/value/" + metricType + "/" + metricName,
	}
}

func metricBuildParamsForUpdate(metricType string, metricName string, metricValue string) params {

	return params{
		metricType:  metricType,
		metricName:  metricName,
		metricValue: metricValue,
		url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
	}
}
