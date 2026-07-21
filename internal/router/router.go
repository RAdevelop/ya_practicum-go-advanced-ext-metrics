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

	r.Post("/update/{metric_type}/{metric_name}/{metric_value}", h.MetricUpdate.ServeHTTP)
	r.Post("/update/{metric_type}/{metric_name}", h.MetricUpdate.ServeHTTP)
	r.Post("/update/{metric_type}", h.MetricUpdate.ServeHTTP)

	return r
}
