package handler

import (
	"encoding/json"
	"fmt"
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

	metric, err := metricGetFromRequest(r)

	if err != nil {
		m.logger.Warn("error", "err", err)
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

	m.metricManager.MetricUpdate(metric)

	if m.config.StoreInterval() != nil && *m.config.StoreInterval() == 0 {
		err = m.metricManager.MetricSnapshotSave()
		if err != nil {
			m.logger.Error("metricManager", "err", err)
		}
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)

	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
		if err != nil {
			m.logger.Warn("error", "err", err)
			http.Error(w, "Can't write response", http.StatusInternalServerError)
		}
	}
}

func (m *Metric) Get(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	metric, err := metricGetFromRequest(r)

	if err != nil {
		m.logger.Warn("error", "err", err)
		http.Error(w, "Can't parse request body", http.StatusBadRequest)
		return
	}

	validatorValue := validator.New()
	validateRes := validateMetricTypeAndName(validatorValue, metric)
	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	metric, err = m.metricManager.MetricValue(metric.MType, metric.ID)
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
		m.logger.Error("error", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

func (m *Metric) List(w http.ResponseWriter, r *http.Request) {

	defer m.requestBodyClose(r)

	var sb strings.Builder
	sb.Grow(1024)

	sb.WriteString("<ul>")

	gaugeMetrics := m.metricManager.MetricList(models.Gauge)
	m.metricListRender(&sb, "Gauge metrics", models.Gauge, gaugeMetrics)

	gaugeMetrics = m.metricManager.MetricList(models.Counter)
	m.metricListRender(&sb, "Counter metrics", models.Counter, gaugeMetrics)

	sb.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		m.logger.Error("error", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

func (m *Metric) metricListRender(sb *strings.Builder, title string, metricType string, metricList map[string]models.Metrics) {
	sb.WriteString("<li><strong>" + title + ":</strong>")

	if len(metricList) > 0 {
		sb.WriteString("<ul>")
		for name, metric := range metricList {
			sb.WriteString("<li>")
			if metricType == models.Gauge {
				sb.WriteString(fmt.Sprintf("%s: %v", name, *metric.Value))
			} else {
				sb.WriteString(fmt.Sprintf("%s: %v", name, *metric.Delta))
			}

			sb.WriteString("</li>")
		}
		sb.WriteString("</ul>")
	}
	sb.WriteString("</li>")
}

func metricGetFromRequest(r *http.Request) (metric *models.Metrics, err error) {

	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":

		if r.Body == nil || r.ContentLength == 0 {
			return nil, fmt.Errorf("body is empty")
		}

		err = json.NewDecoder(r.Body).Decode(&metric)
	default:

		metric = &models.Metrics{}
		metric.MType = r.PathValue("metric_type")
		metric.ID = r.PathValue("metric_name")

		reqMetricValue := r.PathValue("metric_value")

		if reqMetricValue == "" {
			return metric, nil
		}

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

	return metric, err
}

func (m *Metric) requestBodyClose(r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		m.logger.Error("error", "err", err)
	}
}
