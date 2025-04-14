package middleware

import (
	"net/http"
	"time"

	"github.com/kosttiik/pvz-service/internal/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metrics.RequestsTotal.WithLabelValues(r.Method, r.URL.Path).Inc()

		next.ServeHTTP(w, r)

		duration := time.Since(start).Seconds()
		metrics.ResponseTime.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}
