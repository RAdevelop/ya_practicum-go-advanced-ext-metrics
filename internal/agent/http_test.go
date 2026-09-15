package agent

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/agent"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	hServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/router"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/snapshot"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupMockLogger(t *testing.T) *logger.MockLogger {
	logMe := logger.NewMockLogger(t)

	//не знаю как лучше сделать возможное переменное количество параметров для вызова таких методов... :(
	logMe.EXPECT().Info(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Error(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Warn(mock.Anything, mock.Anything, mock.Anything).Maybe()
	logMe.EXPECT().Debug(mock.Anything, mock.Anything, mock.Anything).Maybe()

	return logMe
}

func setupMockConfigProviderServer(t *testing.T) *server.MockConfigProvider {
	cfg := server.NewMockConfigProvider(t)

	cfg.EXPECT().FileStoragePath().Maybe().Return("mock.file")
	cfg.EXPECT().Address().Maybe().Return("localhost:8080")
	cfg.EXPECT().StoreInterval().Maybe().Return(nil)
	cfg.EXPECT().Restore().Maybe().Return(nil)

	return cfg
}
func setupMockConfigProviderAgent(t *testing.T) *agent.MockConfigProvider {
	cfg := agent.NewMockConfigProvider(t)

	cfg.EXPECT().Address().Maybe().Return("localhost:8080")
	cfg.EXPECT().ReportInterval().Maybe().Return(10)
	cfg.EXPECT().PollInterval().Maybe().Return(2)

	return cfg
}

// Тестирование агента (код теста помог написать ИИ)
func TestHttpAgent_Update(t *testing.T) {

	var metricStorage = repository.NewMemory()

	metricSnapshot := snapshot.NewMockAble(t)
	var metricManager = service.NewManager(metricStorage, metricSnapshot)
	logApp := setupMockLogger(t)

	type given struct {
		metrics         models.Metrics
		agentSecretKey  *string
		serverSecretKey *string
	}
	type want struct {
		statusCode int
	}

	tests := []struct {
		name  string
		given given
		want  want
	}{
		{
			name: "gauge StatusOK",
			given: given{
				metrics: models.Metrics{
					MType: "gauge",
					ID:    "test",
					Value: new(42.42),
				},
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "counter StatusOK",
			given: given{
				metrics: models.Metrics{
					MType: "counter",
					ID:    "test",
					Delta: new(int64(42)),
				},
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "gauge WrongType StatusBadRequest",
			given: given{
				metrics: models.Metrics{
					MType: "gaugeWrongType",
					ID:    "test",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "counter WrongType StatusBadRequest",
			given: given{
				metrics: models.Metrics{
					MType: "counterWrongType",
					ID:    "test",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "counter WrongName StatusBadRequest",
			given: given{
				metrics: models.Metrics{
					MType: "counterWrongType",
					ID:    "12WrongName",
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "HashSHA256 sign Ok",
			given: given{
				metrics: models.Metrics{
					MType: "counter",
					ID:    "test",
					Delta: new(int64(42)),
				},
				serverSecretKey: new("abc"),
				agentSecretKey:  new("abc"),
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "HashSHA256 sign Bad",
			given: given{
				metrics: models.Metrics{
					MType: "counter",
					ID:    "test",
					Delta: new(int64(42)),
				},
				serverSecretKey: new("abc"),
				agentSecretKey:  new("cba"),
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfigProviderServer := setupMockConfigProviderServer(t)
			mockConfigProviderAgent := setupMockConfigProviderAgent(t)
			if tt.given.serverSecretKey != nil {
				mockConfigProviderServer.EXPECT().SignKey().Return(*tt.given.serverSecretKey)
			} else {
				mockConfigProviderServer.EXPECT().SignKey().Maybe().Return("")
			}

			if tt.given.agentSecretKey != nil {
				mockConfigProviderAgent.EXPECT().SignKey().Return(*tt.given.agentSecretKey)
			} else {
				mockConfigProviderAgent.EXPECT().SignKey().Maybe().Return("")
			}

			serverContext := &hServer.Context{
				Logger: logApp,
				Config: mockConfigProviderServer,
			}

			h := handler.New(metricManager, serverContext)
			r := router.New(h, serverContext)
			mockServer := httptest.NewServer(r)
			defer mockServer.Close()

			// Создаем resty-клиент с тестовым URL
			client := resty.New()
			client.SetBaseURL(mockServer.URL)
			httpAgent := New(client, mockConfigProviderAgent)

			resp, err := httpAgent.Update(t.Context(), tt.given.metrics)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.statusCode, resp.StatusCode)

			_, err = io.Copy(io.Discard, resp.Body)
			assert.NoError(t, err)
			assert.NoError(t, resp.Body.Close())
		})
	}
}
