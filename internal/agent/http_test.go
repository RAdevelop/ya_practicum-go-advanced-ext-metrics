package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/memory"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/metric"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

var loggerTest = logger.NewTest()

// Тестирование агента (код теста помог написать ИИ)
func TestHttpAgent_Update(t *testing.T) {

	var configProvider = &server.TestConfigProvider{}
	var metricStorage = memory.NewStorage()
	var metricService = metric.NewMetric(metricStorage)
	metricSnapshot, err := snapshot.NewFiler(metricService, configProvider.FileStoragePath())
	assert.NoError(t, err)
	var metricManager = service.NewManager(metricService, metricSnapshot)

	h := handler.New(metricManager, loggerTest, configProvider)
	r := router.New(h)
	mockServer := httptest.NewServer(r)
	defer mockServer.Close()

	// Создаем resty-клиент с тестовым URL
	client := resty.New()
	client.SetBaseURL(mockServer.URL)
	agent := New(client)

	type want struct {
		statusCode int
	}

	tests := []struct {
		name  string
		given models.Metrics
		want  want
	}{
		{
			name: "gauge StatusOK",
			given: models.Metrics{
				MType: "gauge",
				ID:    "test",
				Value: new(42.42),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "counter StatusOK",
			given: models.Metrics{
				MType: "counter",
				ID:    "test",
				Delta: new(int64(42)),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "gauge WrongType StatusBadRequest",
			given: models.Metrics{
				MType: "gaugeWrongType",
				ID:    "test",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "counter WrongType StatusBadRequest",
			given: models.Metrics{
				MType: "counterWrongType",
				ID:    "test",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "counter WrongName StatusBadRequest",
			given: models.Metrics{
				MType: "counterWrongType",
				ID:    "12WrongName",
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := agent.Update(tt.given)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			_, err = io.Copy(io.Discard, resp.Body)
			assert.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
		})
	}
}
