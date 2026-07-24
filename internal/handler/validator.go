package handler

/*
Не стал искать популярные валидаторы для Go.
Чтобы опять же была практика написания именно своего кода.
*/

import (
	"fmt"
	"log"
	"net/http"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

type validateResult struct {
	message    string
	hasError   bool
	httpStatus int
	counter    int64
	gauge      float64
}

func validateMetricTypeAndName(validator *validator.Validator, mType string, mName string) validateResult {
	result := validateMetricType(mType)
	if result.hasError {
		return result
	}

	result = validateMetricName(validator, mName)

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

func validateMetricValue(validator *validator.Validator, mType string, mValue string) validateResult {

	result := validateResult{}

	var err error
	if mType == models.Counter {
		result.counter, err = validator.ValidateValueInt64(mValue)
	} else {
		result.gauge, err = validator.ValidateValueFloat64(mValue)
	}

	if err != nil {
		result.hasError = true
		result.httpStatus = http.StatusBadRequest
		result.message = fmt.Sprintf("Metric value \"%s\" is invalid.", mValue)
	}

	log.Printf("in validateMetricValue(): %+v\n", result)

	return result
}
