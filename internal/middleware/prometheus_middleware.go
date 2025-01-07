package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method, path, and status",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDurationHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "path"},
	)

	// Connection metrics
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// HTTP response size metrics
	httpResponseBytesTotal = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response sizes in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)
)

// PrometheusMiddleware collects HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unmatched" // For 404s
		}

		// Track active connections
		activeConnections.Inc()
		defer activeConnections.Dec()

		c.Next()

		// Record duration
		duration := time.Since(start).Seconds()
		httpRequestDurationHistogram.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		// Record request count
		status := strconv.Itoa(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		// Record response size
		httpResponseBytesTotal.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(float64(c.Writer.Size()))
	}
}

// RegisterMetricsEndpoint adds the /metrics endpoint to the Gin engine
func RegisterMetricsEndpoint(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
