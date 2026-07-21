package agent

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Мок-клиент для тестов
type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

// Тестирование агента (код теста помог написать ИИ)
func TestHttpAgent_Update(t *testing.T) {
	tests := []struct {
		name           string
		metric         MetricIn
		mockResponse   *http.Response
		mockError      error
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "successful update",
			metric: MetricIn{
				Type:  "gauge",
				Name:  "testMetric",
				Value: "42.5",
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`ok`)),
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "server error",
			metric: MetricIn{
				Type:  "wrongType", // StatusBadRequest
				Name:  "testMetric",
				Value: "42.5",
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(bytes.NewBufferString(`error`)),
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  false,
		},
		{
			name: "client error",
			metric: MetricIn{
				Type:  "gauge",
				Name:  "testMetric",
				Value: "42.5",
			},
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок-клиент
			mock := &mockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// Проверяем URL
					expectedURL := "http://localhost:8080/update/" + tt.metric.Type + "/" + tt.metric.Name + "/" + tt.metric.Value
					assert.Equal(t, expectedURL, req.URL.String())

					// Проверяем Content-Type
					assert.Equal(t, "text/plain", req.Header.Get("Content-Type"))

					return tt.mockResponse, tt.mockError
				},
			}

			// Создаем агента с моком
			agent := New(mock)

			// Выполняем запрос
			resp, err := agent.Update(tt.metric)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}
