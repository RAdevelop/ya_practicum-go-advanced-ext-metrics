package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
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

func TestMetric_Update(t *testing.T) {

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

	memStorage := memory.NewStorage()
	metricService := service.NewMetricService(memStorage)
	h := handler.New(metricService)
	r := New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()

	// Создаем resty-клиент с тестовым URL
	client := resty.New()
	client.SetBaseURL(mockServer.URL)

	var result *resty.Response
	var err error

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

func TestMetric_Get(t *testing.T) {

	var metricStorage = memory.NewStorage()
	var metricService = service.NewMetricService(metricStorage)

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
				metricValue: "2",
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

	h := handler.New(metricService)
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

			assert.Equal(t, tt.want.statusCode, result.StatusCode())

			metricValue, err := io.ReadAll(result.RawResponse.Body)
			assert.Equal(t, tt.want.metricValue, string(metricValue))

			assert.NoError(t, err)
			assert.NoError(t, result.RawResponse.Body.Close())
		})
	}
}

func metricBuildParamsForCounterGetValue(metricService *service.MetricService, metricType string, metricName string, metricValues []int64) params {

	for _, value := range metricValues {
		metricService.CounterAdd(metricName, value)
	}

	return params{
		metricType: metricType,
		metricName: metricName,
		url:        "/value/" + metricType + "/" + metricName,
	}
}

func metricBuildParamsForGaugeGetValue(metricService *service.MetricService, metricType string, metricName string, metricValues []float64) params {
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
