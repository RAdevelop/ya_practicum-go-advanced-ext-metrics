package handler

/*
Не стал искать популярные валидаторы для Go.
Чтобы опять же была практика написания именно своего кода.
*/

import (
	"fmt"
	"net/http"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

type validateResult struct {
	message    string
	hasError   bool
	httpStatus int
}

func validateMetricTypeAndName(validator *validator.Validator, metric *models.Metrics) validateResult {
	result := validateMetricType(metric.MType)
	if result.hasError {
		return result
	}

	result = validateMetricName(validator, metric.ID)

	return result
}

func validateMetricType(mType string) validateResult {

	result := validateResult{}

	if mType != models.Counter && mType != models.Gauge {
		result.hasError = true
		result.httpStatus = http.StatusBadRequest

		result.message = fmt.Sprintf("Metric type \"%s\" is not supported.", mType)
		result.message += fmt.Sprintf("\nUse one of the supported metric types: %v", []string{models.Counter, models.Gauge})
	}

	return result
}
func validateMetricName(validator *validator.Validator, mName string) validateResult {

	result := validateResult{}

	if err := validator.ValidateName(mName); err != nil {
		result.hasError = true
		result.message = fmt.Sprintf("Metric name \"%s\" is invalid.", mName)
		result.httpStatus = http.StatusNotFound
	}

	return result
}

// TODO Надо пересмотреть валидацию. Странный валидатор у меня. Если metric установлена, значит Value or Delta прописаны верно.
// Достаточно тогда проверять на nil.
func validateMetricValue(validate *validator.Validator, metric *models.Metrics) validateResult {

	result := validateResult{}

	var err error
	if metric.MType == models.Counter {
		if metric.Delta == nil {
			err = validator.ErrValueInt64
		} else {
			_, err = validate.ValidateValueInt64(metric.StrValue())
		}

	} else {
		if metric.Value == nil {
			err = validator.ErrValueInt64
		} else {
			_, err = validate.ValidateValueFloat64(metric.StrValue())
		}
	}

	if err != nil {
		result.hasError = true
		result.httpStatus = http.StatusBadRequest
		result.message = fmt.Sprintf("Metric value \"%+v\" is invalid.", metric)
	}

	return result
}
