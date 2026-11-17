package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/appcontext"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
)

// fakeManager — минимальная реализация service.MetricManagementAble для примеров.
// Хранит метрики в памяти, чтобы примеры были воспроизводимыми.
type fakeManager struct {
	metrics map[string]models.Metrics
}

func newFakeManager() *fakeManager {
	return &fakeManager{metrics: make(map[string]models.Metrics)}
}

func (f *fakeManager) MetricUpdateBatch(_ context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	for _, m := range metrics {
		f.metrics[m.ID] = m
	}
	return metrics, nil
}

func (f *fakeManager) Metric(_ context.Context, m *models.Metrics) (*models.Metrics, error) {
	stored, ok := f.metrics[m.ID]
	if !ok {
		return nil, fmt.Errorf("metric %q not found", m.ID)
	}
	return &stored, nil
}

func (f *fakeManager) MetricList(_ context.Context, metricType string) ([]models.Metrics, error) {
	out := make([]models.Metrics, 0, len(f.metrics))
	for _, m := range f.metrics {
		if m.MType == metricType {
			out = append(out, m)
		}
	}
	return out, nil
}

func (f *fakeManager) MetricSnapshotSave(context.Context) error { return nil }
func (f *fakeManager) StoragePing(context.Context) error        { return nil }
func (f *fakeManager) MetricSnapshotLoad(context.Context) error { return nil }

// newTestAppContext — контекст приложения для тестов:
// реальный логгер + реальный конфиг сервера с интервалом сохранения больше нуля.
func newTestAppContext() *appcontext.AppContext {
	env, err := configServer.NewEnv()
	if err != nil {
		panic(err) // env-переменные в тестовом окружении обычно не заданы, NewEnv вернёт дефолты
	}

	cfg := configServer.New(env)

	// Задаем интервал явно, чтобы в хендлере Update не срабатывал
	// путь с синхронным MetricSnapshotSave (он нужен только при interval == 0).
	interval := uint(10)
	cfg.StoreIntervalSet(&interval)

	return &appcontext.AppContext{
		Logger: logger.New(),
		Config: cfg,
	}
}

// newTestServer — поднимает сервер со всеми эндпоинтами из практического трека.
func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	m := handler.NewMetric(newFakeManager(), newTestAppContext())

	mux.HandleFunc("POST /update/{metric_type}/{metric_name}/{metric_value}", m.Update)
	mux.HandleFunc("POST /update", m.Update)
	mux.HandleFunc("POST /updates/", m.UpdateBatch)
	mux.HandleFunc("POST /value", m.Get)
	mux.HandleFunc("GET /", m.List)
	mux.HandleFunc("GET /ping", m.StoragePing)

	return httptest.NewServer(mux)
}

// ExampleMetric_Update демонстрирует обновление метрики в JSON-формате:
// POST /update с Content-Type: application/json.
func ExampleMetric_Update() {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal(models.Metrics{
		ID:    "PollCount",
		MType: models.Counter,
		Delta: int64ptr(10),
	})

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	var got models.Metrics
	_ = json.NewDecoder(resp.Body).Decode(&got)
	fmt.Println(got.ID, got.MType, *got.Delta)

	// Output:
	// 200
	// PollCount counter 10
}

// ExampleMetric_update демонстрирует обновление метрики в текстовом формате:
// POST /update/{metric_type}/{metric_name}/{metric_value}.
func ExampleMetric_update() {
	srv := newTestServer()
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/update/gauge/Alloc/123.45", "text/plain", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	// Output:
	// 200
}

// ExampleMetric_UpdateBatch демонстрирует пакетное обновление:
// POST /updates/ с массивом метрик.
func ExampleMetric_UpdateBatch() {
	srv := newTestServer()
	defer srv.Close()

	body, _ := json.Marshal([]models.Metrics{
		{ID: "Requests", MType: models.Counter, Delta: int64ptr(1)},
		{ID: "Alloc", MType: models.Gauge, Value: float64ptr(3.14)},
	})

	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/updates/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	var got []models.Metrics
	_ = json.NewDecoder(resp.Body).Decode(&got)
	fmt.Println(len(got))

	// Output:
	// 200
	// 2
}

// ExampleMetric_Get демонстрирует чтение значения метрики: POST /value.
func ExampleMetric_Get() {
	srv := newTestServer()
	defer srv.Close()

	// сначала записываем метрику
	body, _ := json.Marshal(models.Metrics{
		ID: "PollCount", MType: models.Counter, Delta: int64ptr(10),
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	// затем читаем её
	body, _ = json.Marshal(models.Metrics{ID: "PollCount", MType: models.Counter})
	req, _ = http.NewRequest(http.MethodPost, srv.URL+"/value", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	var got models.Metrics
	_ = json.NewDecoder(resp.Body).Decode(&got)
	fmt.Println(got.ID, got.MType, *got.Delta)

	// Output:
	// 200
	// PollCount counter 10
}

// ExampleMetric_List демонстрирует получение HTML-списка всех метрик: GET /.
func ExampleMetric_List() {
	srv := newTestServer()
	defer srv.Close()

	// записываем метрику, чтобы она попала в список
	body, _ := json.Marshal(models.Metrics{
		ID: "PollCount", MType: models.Counter, Delta: int64ptr(10),
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()

	resp, err = http.Get(srv.URL + "/")
	if err != nil {
		return
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	fmt.Println(strings.Contains(buf.String(), "PollCount: 10"))

	// Output:
	// 200
	// true
}

// вспомогательные конструкторы указателей
func int64ptr(v int64) *int64       { return &v }
func float64ptr(v float64) *float64 { return &v }
