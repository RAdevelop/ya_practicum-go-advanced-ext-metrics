package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/converter"
	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/service"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

type Metric struct {
	metricService *service.MetricService
}

func NewMetric(metricService *service.MetricService) *Metric {
	return &Metric{
		metricService: metricService,
	}
}

/*
Update - обновляем данные по метрикам

Формат url pth: /{metric_type}/{metric_name}/{metric_value}
*/
func (m *Metric) Update(w http.ResponseWriter, r *http.Request) {

	metricType := r.PathValue("metric_type")
	metricName := r.PathValue("metric_name")
	metricValue := r.PathValue("metric_value")

	validatorValue := validator.New()
	validateRes := validateMetricTypeAndName(validatorValue, metricType, metricName)
	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	validateRes = validateMetricValue(validatorValue, metricType, metricValue)

	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	if metricType == models.Counter {
		m.metricService.CounterAdd(metricName, validateRes.counter)
	} else {
		m.metricService.GaugeUpdate(metricName, validateRes.gauge)
	}

	w.WriteHeader(http.StatusOK)
}

func (m *Metric) Get(w http.ResponseWriter, r *http.Request) {

	metricType := r.PathValue("metric_type")
	metricName := r.PathValue("metric_name")

	validatorValue := validator.New()
	validateRes := validateMetricTypeAndName(validatorValue, metricType, metricName)
	if validateRes.hasError {
		http.Error(w, validateRes.message, validateRes.httpStatus)
		return
	}

	var metricValue any
	var err error

	if metricType == models.Counter {
		metricValue, err = m.metricService.CounterByNameAccumulative(metricName)
	} else {
		metricValue, err = m.metricService.GaugeByName(metricName)
	}

	if err != nil {
		http.Error(w, "Metric value not found by name", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(converter.NumericToString(metricValue)))
	if err != nil {
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}

func (m *Metric) List(w http.ResponseWriter, r *http.Request) {

	var sb strings.Builder
	sb.Grow(1024)

	sb.WriteString("<ul>")
	sb.WriteString("<li><strong>Gauge metrics:</strong></li>")

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
	sb.WriteString("<li><strong>Counter metrics:</strong></li>")
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

	sb.WriteString("</ul>")

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}
