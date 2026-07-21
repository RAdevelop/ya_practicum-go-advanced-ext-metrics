package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("counter", "someMetric", "527"),
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("counter", "someMetric", "527"),
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("counter", "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "counter metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("unknownMetricType" /*причина StatusBadRequest*/, "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("123counterBadName" /*причина StatusBadRequest*/, "someMetric", "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "counter metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("counter", "123someMetricBadName" /*причина StatusNotFound*/, "527"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name: "counter metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("counter", "someMetric", "s527" /*причина StatusBadRequest*/),
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("gauge", "someMetric", "527.123"),
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("gauge", "someMetric", "527.123"),
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
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("gauge", "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusUnsupportedMediaType,
			},
		},
		{
			name: "gauge metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("unknownMetricType" /*причина StatusBadRequest*/, "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "gauge metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("123gaugeBadName" /*причина StatusBadRequest*/, "someMetric", "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
		{
			name: "gauge metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("gauge", "123someMetricBadName" /*причина StatusNotFound*/, "527.123"),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusNotFound,
			},
		},
		{
			name: "gauge metric update with StatusUnsupportedMediaType",
			given: given{
				contentType: "text/plain",
				method:      http.MethodPost,
				reqParams: (func(metricType string, metricName string, metricValue string) params {
					return params{
						metricType:  metricType,
						metricName:  metricName,
						metricValue: metricValue,
						url:         "/update/" + metricType + "/" + metricName + "/" + metricValue,
					}
				})("gauge", "someMetric", "s527.123" /*причина StatusBadRequest*/),
			},
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusBadRequest,
			},
		},
	}

	h := handler.New()
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

			assert.Equal(t, tt.want.statusCode, result.StatusCode())
			// io.Discard выступает в качестве приёмника ненужных данных
			_, err = io.Copy(io.Discard, result.RawResponse.Body)
			assert.NoError(t, err)
			assert.NoError(t, result.RawResponse.Body.Close())
		})
	}
}
