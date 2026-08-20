package router

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handler.Handlers) http.Handler {

	r := chi.NewRouter()

	r.Route("/update", func(r chi.Router) {

		r.With(middleware.AllowContentType("application/json")).
			Post("/", h.MetricUpdate.ServeHTTP)

		r.Route("/{metric_type}", func(r chi.Router) {
			r.Use(middleware.AllowContentType("text/plain"))

			r.Post("/", h.MetricUpdate.ServeHTTP)
			r.Post("/{metric_name}", h.MetricUpdate.ServeHTTP)
			r.Post("/{metric_name}/{metric_value}", h.MetricUpdate.ServeHTTP)
		})
	})

	r.Route("/value", func(r chi.Router) {
		r.With(middleware.AllowContentType("application/json")).Post("/", h.MetricGet.ServeHTTP)
		r.With(middleware.AllowContentType("text/plain")).Get("/{metric_type}/{metric_name}", h.MetricGet.ServeHTTP)
	})

	r.Get("/", h.MetricList.ServeHTTP)

	return r
}
