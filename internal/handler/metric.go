package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/appcontext"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service/audit"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

// Metric — набор HTTP-обработчиков для работы с метриками.
//
// Поддерживает типы counter и gauge и два формата обмена:
// text/plain (URL-параметры) и application/json.
type Metric struct {
	metricManager service.MetricManagementAble
	appContext    *appcontext.AppContext
}

// NewMetric — конструктор обработчика метрик.
//
// Принимает менеджер метрик (сервисный слой) и контекст приложения
// (логгер, конфигурация). Возвращает обработчик, готовый к регистрации в роутере.
func NewMetric(metricManager service.MetricManagementAble, appContext *appcontext.AppContext) *Metric {
	return &Metric{
		metricManager: metricManager,
		appContext:    appContext,
	}
}

// Update — обработчик обновления одной метрики.
//
// Поддерживает два формата запроса.
//
// Формат text/plain (URL-параметры):
//
//	POST /update/{metric_type}/{metric_name}/{metric_value}
//	Пример: /update/gauge/Alloc/123.45 или /update/counter/PollCount/10
//
// Формат application/json:
//
//	POST /update
//	{"id":"PollCount","type":"counter","value":10}
//	{"id":"Alloc","type":"gauge","value":123.123}
//
// Ответ: 200 OK; для JSON-запроса возвращает обновлённую метрику в теле.
// Ошибки: 400 — невалидный запрос или ошибка сохранения, 500 — ошибка записи ответа.
func (m *Metric) Update(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metrics, err := metricsGetFromRequest(r)
	metric := &models.Metrics{}
	if len(metrics) == 1 && err == nil {
		metric = &metrics[0]
	}

	if err != nil {
		m.appContext.Logger.Error("HandlerMetricUpdate", "err", err)
		http.Error(w, "Can't parse request body", http.StatusBadRequest)
		return
	}

	validatorValue := validator.New()
	validateRes := validateMetricTypeAndName(validatorValue, metric)
	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	validateRes = validateMetricValue(validatorValue, metric)

	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	ctx := audit.WithIP(r.Context(), clientIP(r))

	if _, err = m.metricManager.MetricUpdateBatch(ctx, []models.Metrics{*metric}); err != nil {
		m.appContext.Logger.Warn("HandlerMetricUpdate", "err", err)
		http.Error(w, "Can't update metric", http.StatusBadRequest)
		return
	}

	if m.appContext.Config.StoreInterval() != nil && *m.appContext.Config.StoreInterval() == 0 {
		err = m.metricManager.MetricSnapshotSave(ctx)
		if err != nil {
			m.appContext.Logger.Error("HandlerMetricUpdate", "err", err)
		}
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			m.appContext.Logger.Warn("HandlerMetricUpdate", "err", err)
			http.Error(w, "Can't write response", http.StatusInternalServerError)
		}
	}
}

// UpdateBatch — обработчик пакетного обновления метрик.
//
//	POST /updates/
//	Content-Type: application/json
//	Тело: массив метрик [{"id":"A","type":"counter","value":1}, ...]
//
// Все метрики валидируются до сохранения; при любой ошибке валидации
// батч целиком отклоняется с кодом 400.
// Ответ: 200 OK со списком обновлённых метрик (для JSON-запроса).
func (m *Metric) UpdateBatch(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metrics, err := metricsGetFromRequest(r)

	if err != nil {
		m.appContext.Logger.Error("HandlerMetricUpdateBatch", "err", err)
		http.Error(w, "Can't parse request body", http.StatusBadRequest)
		return
	}

	validatorValue := validator.New()

	// копим ошибки
	var validateErrors []error
	for _, metric := range metrics {
		validateRes := validateMetricTypeAndName(validatorValue, &metric)
		if validateRes.hasError {
			validateErrors = append(validateErrors, fmt.Errorf("metric validation type and name failed: %s", validateRes.message))
		}

		validateRes = validateMetricValue(validatorValue, &metric)
		if validateRes.hasError {
			validateErrors = append(validateErrors, fmt.Errorf("metric validation value failed: %s", validateRes.message))
		}
	}

	if len(validateErrors) > 0 {
		m.appContext.Logger.Error("HandlerMetricUpdateBatch", "err", errors.Join(validateErrors...))
		http.Error(w, "Metric validation failed", http.StatusBadRequest)
		return
	}

	ctx := audit.WithIP(r.Context(), clientIP(r))

	if metrics, err = m.metricManager.MetricUpdateBatch(ctx, metrics); err != nil {
		m.appContext.Logger.Error("HandlerMetricUpdateBatch", "err", err)
		http.Error(w, "Can't update metric", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metrics)
		if err != nil {
			m.appContext.Logger.Warn("HandlerMetricUpdateBatch", "err", err)
			http.Error(w, "Can't write response", http.StatusInternalServerError)
		}
	}
}

// Get — обработчик получения значения метрики.
//
// Формат text/plain:
//
//	POST|GET /value/{metric_type}/{metric_name}
//
// Формат application/json:
//
//	POST /value
//	{"id":"PollCount","type":"counter"}
//
// Ответ: 200 OK со значением метрики (JSON или строка).
// Ошибки: 404 — метрика не найдена, 400 — невалидный запрос.
func (m *Metric) Get(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metrics, err := metricsGetFromRequest(r)

	if err != nil {
		m.appContext.Logger.Warn("error", "err", err)
		http.Error(w, "Can't parse request body", http.StatusBadRequest)
		return
	}

	metric := &models.Metrics{}
	if len(metrics) == 1 {
		metric = &metrics[0]
	}

	validatorValue := validator.New()
	validateRes := validateMetricTypeAndName(validatorValue, metric)
	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	metric, err = m.metricManager.Metric(r.Context(), metric)
	if err != nil {
		m.appContext.Logger.Error("HandlerMetricGet", "err", err)
		http.Error(w, "Metric value not found by name", http.StatusNotFound)
		return
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
	} else {
		_, err = w.Write([]byte(metric.StrValue()))
	}

	if err != nil {
		m.appContext.Logger.Error("HandlerMetricGet", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

// List — обработчик списка всех метрик.
//
//	GET /
//
// Возвращает HTML-страницу со списком gauge- и counter-метрик.
func (m *Metric) List(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	var sb strings.Builder
	sb.Grow(1024)

	sb.WriteString("<ul>")

	gaugeMetrics, err := m.metricManager.MetricList(r.Context(), models.Gauge)
	if err != nil {
		m.appContext.Logger.Error("HandlerMetricList", "err", err)
	} else {
		m.metricListRender(&sb, "Gauge metrics", models.Gauge, gaugeMetrics)
	}

	gaugeMetrics, err = m.metricManager.MetricList(r.Context(), models.Counter)
	if err != nil {
		m.appContext.Logger.Error("HandlerMetricList", "err", err)
	} else {
		m.metricListRender(&sb, "Counter metrics", models.Counter, gaugeMetrics)
	}

	sb.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		m.appContext.Logger.Error("HandlerMetricList", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

// StoragePing — обработчик проверки доступности хранилища.
//
//	GET /ping
//
// Ответ: 200 OK, если соединение установлено; 500 — если недоступно.
func (m *Metric) StoragePing(w http.ResponseWriter, r *http.Request) {
	defer m.requestBodyClose(r)

	err := m.metricManager.StoragePing(r.Context())

	if err != nil {
		m.appContext.Logger.Error("StoragePing", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// metricsGetFromRequest извлекает метрики из запроса.
//
// Поддерживает JSON (одна метрика или массив) и URL-параметры
// /{metric_type}/{metric_name}/{metric_value}.
func (m *Metric) metricListRender(sb *strings.Builder, title string, metricType string, metricList []models.Metrics) {
	sb.WriteString("<li><strong>" + title + ":</strong>")

	if len(metricList) > 0 {
		sb.WriteString("<ul>")
		for _, metric := range metricList {
			sb.WriteString("<li>")
			if metricType == models.Gauge {
				sb.WriteString(fmt.Sprintf("%s: %v", metric.ID, *metric.Value))
			} else {
				sb.WriteString(fmt.Sprintf("%s: %v", metric.ID, *metric.Delta))
			}

			sb.WriteString("</li>")
		}
		sb.WriteString("</ul>")
	}
	sb.WriteString("</li>")
}

func metricsGetFromRequest(r *http.Request) (metrics []models.Metrics, err error) {

	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":

		if r.Body == nil || r.ContentLength == 0 {
			return nil, fmt.Errorf("body is empty")
		}
		var reqBody []byte
		reqBody, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}

		err = json.Unmarshal(reqBody, &metrics)
		if err == nil {
			return metrics, nil
		}

		if _, ok := errors.AsType[*json.UnmarshalTypeError](err); ok {
			metric := &models.Metrics{}
			err = json.Unmarshal(reqBody, metric)
			if err != nil {
				return nil, fmt.Errorf("decode single: %w", err)
			}

			metrics = []models.Metrics{*metric}
		}

	default:

		metric := &models.Metrics{}
		metric.MType = r.PathValue("metric_type")
		metric.ID = r.PathValue("metric_name")

		reqMetricValue := r.PathValue("metric_value")

		if reqMetricValue != "" {
			switch metric.MType {
			case models.Gauge:
				var mV float64
				mV, err = converter.ToFloat64(reqMetricValue)
				metric.Value = &mV
			case models.Counter:
				var mV int64
				mV, err = strconv.ParseInt(reqMetricValue, 10, 64)
				metric.Delta = &mV
			}
		}

		metrics = []models.Metrics{*metric}
	}

	return metrics, err
}

func (m *Metric) requestBodyClose(r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		m.appContext.Logger.Error("error", "err", err)
	}
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
