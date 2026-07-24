package handler

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/validator"
)

/*
Для информации.
Данный код добавлен для практики реализации и работы с Middleware
*/

type Middleware func(http.Handler) http.Handler

/*
MiddlewarePipeLine - пайплан обработки запросов, используя список "Middleware"

Важно: MiddlewarePipeLine(handler, middleware1, middleware2, ..., middlewareN)
мидлвари будут применяться в обратном порядке их добавления, то есть:
сначала middlewareN, потом middlewareN-1 и так далее
*/
func MiddlewarePipeLine(h http.Handler, middlewares ...Middleware) http.Handler {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func MiddlewareIsPostRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// MiddlewareValidator - проверка параметров запроса
func MiddlewareValidator(next http.Handler) http.Handler {
	// получаем Handler приведением типа http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		metricType := r.PathValue("metric_type")
		metricName := r.PathValue("metric_name")
		metricValue := r.PathValue("metric_value")

		validatorValue := validator.New()

		validateRes := validateMetricTypeAndName(validatorValue, metricType, metricName)
		if validateRes.hasError {
			http.Error(w, validateRes.message, validateRes.httpStatus)
			return
		}

		if r.Method == http.MethodPost {
			validateRes = validateMetricValue(validatorValue, metricType, metricValue)
			if validateRes.hasError {
				// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest.
				http.Error(w, validateRes.message, validateRes.httpStatus)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
