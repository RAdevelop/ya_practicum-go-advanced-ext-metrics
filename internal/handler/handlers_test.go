package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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

	handlers := New()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.given.method, tt.given.reqParams.url, nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.given.contentType)
			req.SetPathValue("metric_type", tt.given.reqParams.metricType)
			req.SetPathValue("metric_name", tt.given.reqParams.metricName)
			req.SetPathValue("metric_value", tt.given.reqParams.metricValue)

			resWriter := httptest.NewRecorder()

			handlers.MetricUpdate.ServeHTTP(resWriter, req)

			result := resWriter.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			// io.Discard выступает в качестве приёмника ненужных данных
			_, err = io.Copy(io.Discard, result.Body)
			assert.NoError(t, err)
			assert.NoError(t, result.Body.Close())
		})
	}
}
