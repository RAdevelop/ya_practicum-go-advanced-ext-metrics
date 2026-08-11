package handler

import (
	"net/http"
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
