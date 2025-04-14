package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)

	ResponseTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	PvzCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of PVZ created",
		},
	)

	OrderReceiptsCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "order_receipts_created_total",
			Help: "Total number of order receipts created",
		},
	)

	ProductsAddedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of products added",
		},
	)
)
