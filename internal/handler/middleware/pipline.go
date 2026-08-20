package middleware

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
)

/*
Для информации.
Данный код добавлен для практики реализации и работы с Middleware
*/

type Middleware func(logger.Logger, http.Handler) http.Handler

/*
PipeLine - поток обработки запросов, используя список "Middleware"
*/
func PipeLine(logger logger.Logger, h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](logger, h)
	}
	return h
}
