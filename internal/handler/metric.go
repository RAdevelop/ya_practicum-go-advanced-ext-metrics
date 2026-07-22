package handler

import (
	"fmt"
	"net/http"

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
