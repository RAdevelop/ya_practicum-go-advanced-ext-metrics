package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	configServer "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/config/server"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

type Metric struct {
	metricManager service.MetricManagementAble
	logger        logger.Logger
	config        configServer.ConfigProvider
}

func NewMetric(metricManager service.MetricManagementAble, logger logger.Logger, config configServer.ConfigProvider) *Metric {
	return &Metric{
		metricManager: metricManager,
		logger:        logger,
		config:        config,
	}
}

/*
Update - обновляем данные по метрикам

Формат GET:
- "Content-Type": "text/plain"
- url: /update/{metric_type}/{metric_name}/{metric_value}
Формат POST:
- "Content-Type": "application/json"
- url: /update
- body:
  - counter: {"id": "metricName","type": "counter","value": 123}
  - gauge: {"id": "metricName","type": "gauge","value": 123.123}
*/
func (m *Metric) Update(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metrics, err := metricsGetFromRequest(r)
	metric := &models.Metrics{}
	if len(metrics) == 1 && err == nil {
		metric = &metrics[0]
	}

	if err != nil {
		m.logger.Error("HandlerMetricUpdate", "err", err)
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

	if err = m.metricManager.MetricUpdate(r.Context(), metric); err != nil {
		m.logger.Warn("HandlerMetricUpdate", "err", err)
		http.Error(w, "Can't update metric", http.StatusBadRequest)
		return
	}

	if m.config.StoreInterval() != nil && *m.config.StoreInterval() == 0 {
		err = m.metricManager.MetricSnapshotSave(r.Context())
		if err != nil {
			m.logger.Error("HandlerMetricUpdate", "err", err)
		}
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			m.logger.Warn("HandlerMetricUpdate", "err", err)
			http.Error(w, "Can't write response", http.StatusInternalServerError)
		}
	}
}

// UpdateBatch - обновляем метрики пачкой
// TODO UpdateBatch
func (m *Metric) UpdateBatch(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metric, err := metricGetFromRequest(r)

	if err != nil {
		m.logger.Error("HandlerMetricUpdateBatch", "err", err)
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

	if err = m.metricManager.MetricUpdate(r.Context(), metric); err != nil {
		m.logger.Warn("HandlerMetricUpdateBatch", "err", err)
		http.Error(w, "Can't update metric", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			m.logger.Warn("HandlerMetricUpdateBatch", "err", err)
			http.Error(w, "Can't write response", http.StatusInternalServerError)
		}
	}
}

func (m *Metric) Get(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metrics, err := metricsGetFromRequest(r)

	if err != nil {
		m.logger.Warn("error", "err", err)
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

	metric, err = m.metricManager.MetricValue(r.Context(), metric.MType, metric.ID)
	if err != nil {
		m.logger.Warn("error", "err", err)
		http.Error(w, "Metric value not found by name", http.StatusNotFound)
		return
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
	} else {

		var metricValue any
		switch metric.MType {
		case models.Gauge:
			metricValue = *metric.Value
		case models.Counter:
			metricValue = *metric.Delta
		}

		_, err = w.Write([]byte(converter.NumericToString(metricValue)))
	}

	if err != nil {
		m.logger.Error("HandlerMetricGet", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

func (m *Metric) List(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	var sb strings.Builder
	sb.Grow(1024)

	sb.WriteString("<ul>")

	gaugeMetrics, err := m.metricManager.MetricList(r.Context(), models.Gauge)
	if err != nil {
		m.logger.Error("HandlerMetricList", "err", err)
	} else {
		m.metricListRender(&sb, "Gauge metrics", models.Gauge, gaugeMetrics)
	}

	gaugeMetrics, err = m.metricManager.MetricList(r.Context(), models.Counter)
	if err != nil {
		m.logger.Error("HandlerMetricList", "err", err)
	} else {
		m.metricListRender(&sb, "Counter metrics", models.Counter, gaugeMetrics)
	}

	sb.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(sb.String()))
	if err != nil {
		m.logger.Error("HandlerMetricList", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

func (m *Metric) StoragePing(w http.ResponseWriter, r *http.Request) {
	defer m.requestBodyClose(r)

	err := m.metricManager.StoragePing(r.Context())

	if err != nil {
		m.logger.Error("StoragePing", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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
		m.logger.Error("error", "err", err)
	}
}
