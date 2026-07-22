package router

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handler.Handlers) http.Handler {

	r := chi.NewRouter()

	r.Use(middleware.AllowContentType("text/plain"))

	r.Route("/update", func(r chi.Router) {
		r.Post("/{metric_type}/{metric_name}/{metric_value}", h.MetricUpdate.ServeHTTP)
		r.Post("/{metric_type}/{metric_name}", h.MetricUpdate.ServeHTTP)
		r.Post("/{metric_type}", h.MetricUpdate.ServeHTTP)
	})

	r.Route("/value", func(r chi.Router) {
		r.Get("/{metric_type}/{metric_name}", h.MetricGet.ServeHTTP)
	})

	return r
}
