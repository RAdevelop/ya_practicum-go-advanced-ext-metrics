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
	storage service.MetricStorage
}

func NewMetric(storage service.MetricStorage) *Metric {
	return &Metric{
		storage: storage,
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
	switch metricType {
	case models.Counter:

		mValue, _ := validateValue.ValidateValueInt64(metricValue)
		m.storage.CounterAdd(metricName, mValue)

	case models.Gauge:
		mValue, _ := validateValue.ValidateValueFloat64(metricValue)
		m.storage.GaugeUpdate(metricName, mValue)
	}

	w.WriteHeader(http.StatusOK)

	var sb strings.Builder
	sb.Grow(256)

	sb.WriteString(fmt.Sprintf("metricType: %s", metricType))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("metricName: %s", metricName))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("metricValue: %s", metricValue))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("memStorage.Gauge(): %v", memStorage.Gauge()))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("memStorage.Counter(): %v\n", memStorage.Counter()))
	sb.WriteString("\n")

	_, err := w.Write([]byte(sb.String()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
