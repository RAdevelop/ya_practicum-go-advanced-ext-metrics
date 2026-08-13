package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

type Metric struct {
	metricService *service.MetricService
	logger        logger.Logger
}

func NewMetric(metricService *service.MetricService, logger logger.Logger) *Metric {
	return &Metric{
		metricService: metricService,
		logger:        logger,
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

	if metric.MType == models.Counter {
		m.metricService.CounterAdd(metric.ID, *metric.Delta)
	} else {
		m.metricService.GaugeUpdate(metric.ID, *metric.Value)
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

	var metricValue any

	if metric.MType == models.Counter {
		metricValue, err = m.metricService.CounterByNameAccumulative(metric.ID)
	} else {
		metricValue, err = m.metricService.GaugeByName(metric.ID)
	}

	if err != nil {
		m.logger.Warn("error", "err", err)
		http.Error(w, "Metric value not found by name", http.StatusNotFound)
		return
	}

	switch mValue := metricValue.(type) {
	case float64:
		metric.Value = &mValue
	case int64:
		metric.Delta = &mValue
	}

	contentType := r.Header.Get("Content-Type")
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	if contentType == "application/json" {
		err = json.NewEncoder(w).Encode(metric)
	} else {
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
	sb.WriteString("<li><strong>Gauge metrics:</strong>")

	gaugeMetrics := m.metricService.Gauge()
	if len(gaugeMetrics) > 0 {
		sb.WriteString("<ul>")
		for name, value := range gaugeMetrics {
			// Порядок метрик при выводе на страницу будет чаще всего разный всегда из-за перебора по "map"
			// Но прям требования на этот счет не было. Поэтому оставляю пока так.
			sb.WriteString("<li>")
			sb.WriteString(fmt.Sprintf("%s: %v", name, value))
			sb.WriteString("</li>")
		}
		sb.WriteString("</ul>")
	}
	sb.WriteString("</li>")

	sb.WriteString("<li><strong>Counter metrics:</strong>")
	counterMetrics := m.metricService.CounterAccumulative()
	if len(counterMetrics) > 0 {
		sb.WriteString("<ul>")
		for name, value := range counterMetrics {
			sb.WriteString("<li>")
			sb.WriteString(fmt.Sprintf("%s: %v", name, value))
			sb.WriteString("</li>")
		}
		sb.WriteString("</ul>")
	}
	sb.WriteString("</li>")
	sb.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		m.logger.Error("error", "err", err)
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
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
