package middleware

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/server"
)

/*
Для информации.
Данный код добавлен для практики реализации и работы с Middleware
*/

type Middleware func(*server.Context, http.Handler) http.Handler

/*
PipeLine - поток обработки запросов, используя список "Middleware"
*/
func PipeLine(serverContext *server.Context, h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](serverContext, h)
	}
	return h
}
