package router

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	appMiddleware "github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler/middleware"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handler.Handlers, logger logger.Logger) http.Handler {

	r := chi.NewRouter()

	r.NotFound(appMiddleware.WithLogging(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})).ServeHTTP)

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

	r.With(middleware.AllowContentType("application/json")).Post("/updates/", h.MetricUpdateBatch.ServeHTTP)

	r.Route("/value", func(r chi.Router) {
		r.With(middleware.AllowContentType("application/json")).Post("/", h.MetricGet.ServeHTTP)
		r.With(middleware.AllowContentType("text/plain")).Get("/{metric_type}/{metric_name}", h.MetricGet.ServeHTTP)
	})

	r.With(middleware.AllowContentType("text/plain")).Get("/ping", h.MetricStoragePing.ServeHTTP)

	r.Get("/", h.MetricList.ServeHTTP)

	return r
}
