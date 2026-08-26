package router

import (
	"net/http"

	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/handler"
	"github.com/RAdevelop/ya_practicum-go-advanced-ext-metrics/internal/repository/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func New(h *handler.Handlers, db database.Database) http.Handler {

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

	/*
		Пока проверка доступности БД сделаю тут.
		Если хранилище метрик из "memory" переедет в "БД", тогда перенести в хэндлеры для метрик
	*/
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {

		err := db.Ping()
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/", h.MetricList.ServeHTTP)

	return r
}
