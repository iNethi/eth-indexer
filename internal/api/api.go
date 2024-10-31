package api

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-chi/chi/v5"
)

func New() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/metrics", metricsHandler())

	return r
}

func metricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		metrics.WritePrometheus(w, true)
	}
}
