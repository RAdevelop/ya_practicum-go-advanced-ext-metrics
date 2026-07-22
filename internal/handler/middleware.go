package handler

import (
	"fmt"
	"net/http"

	models "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/model"
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

func MiddlewareContentTypeTextPlain(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
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

		// При попытке передать запрос с некорректным типом метрики возвращать http.StatusBadRequest.
		if metricType != models.Counter && metricType != models.Gauge {

			errMsg := fmt.Sprintf("Metric type \"%s\" is not supported.", metricType)
			errMsg += fmt.Sprintf("\nUse one of the supported metric types: %v", []string{models.Counter, models.Gauge})

			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}

		validatorValue := validator.New()
		var err error

		// При попытке передать запрос без имени метрики возвращать http.StatusNotFound.
		if err = validatorValue.ValidateName(metricName); err != nil {
			errMsg := fmt.Sprintf("Metric name \"%s\" is invalid.", metricName)
			http.Error(w, errMsg, http.StatusNotFound)
			return
		}

		if r.Method == http.MethodPost {
			switch metricType {
			case models.Counter:
				_, err = validatorValue.ValidateValueInt64(metricValue)
			case models.Gauge:
				_, err = validatorValue.ValidateValueFloat64(metricValue)

			}

			// При попытке передать запрос с некорректным значением возвращать http.StatusBadRequest.
			if err != nil {
				errMsg := fmt.Sprintf("Metric value \"%s\" is invalid.", metricValue)
				http.Error(w, errMsg, http.StatusBadRequest)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
