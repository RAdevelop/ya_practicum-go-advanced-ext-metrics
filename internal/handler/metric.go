package handler

import (
	"fmt"
	"net/http"
	"strings"

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

	validateValue := validator.New()
	/*
		Здесь конечно же надо проверять тоже параметры запроса для метрик.
		Так как middleware с валидацией может быть как добавлен, так и удален из цепочки выполнения запроса.
		И если его не будет, то данный обработчик обновления метрик будет работать не корректно.
		Например, если покрыть тестами только данный обработчик, его не корректная работа сразу проявится.
		Это не сделано, для практики реализации цепочки middleware
	*/
	switch metricType {
	case models.Counter:

		mValue, _ := validateValue.ValidateValueInt64(metricValue)
		m.metricService.CounterAdd(metricName, mValue)

	case models.Gauge:
		mValue, _ := validateValue.ValidateValueFloat64(metricValue)
		m.metricService.GaugeUpdate(metricName, mValue)
	default:
		http.Error(w, "Metric type not supported", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (m *Metric) Get(w http.ResponseWriter, r *http.Request) {

	metricType := r.PathValue("metric_type")
	metricName := r.PathValue("metric_name")

	var metricValue any
	var err error

	switch metricType {
	case models.Counter:
		metricValue, err = m.metricService.CounterByNameAccumulative(metricName)
	case models.Gauge:
		metricValue, err = m.metricService.GaugeByName(metricName)
	default:
		http.Error(w, "Metric type not supported", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, "Metric value not found by name", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(fmt.Sprintf("%v", metricValue)))
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
	counterMetrics := m.metricService.Counter()
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
	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		http.Error(w, "Can't write response", http.StatusInternalServerError)
	}
}
