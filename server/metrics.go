package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of HTTP requests processed, labeled by path and method.",
		},
		[]string{"path", "method"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of request durations by path and method.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

func instrumentHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timer := prometheus.NewTimer(httpRequestDuration.WithLabelValues(r.URL.Path, r.Method))
		defer timer.ObserveDuration()

		httpRequestsTotal.WithLabelValues(r.URL.Path, r.Method).Inc()

		next.ServeHTTP(w, r)
	})
}
