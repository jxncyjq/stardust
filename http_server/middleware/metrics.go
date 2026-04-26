package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge
	metricsOnce          sync.Once
)

func initMetrics(serviceName string) {
	metricsOnce.Do(func() {
		httpRequestsTotal = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: serviceName,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		)
		httpRequestDuration = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: serviceName,
				Name:      "http_request_duration_ms",
				Help:      "HTTP request duration in milliseconds",
				Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
			},
			[]string{"method", "path"},
		)
		httpRequestsInFlight = prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: serviceName,
				Name:      "http_requests_in_flight",
				Help:      "Number of HTTP requests currently being processed",
			},
		)
		prometheus.MustRegister(httpRequestsTotal, httpRequestDuration, httpRequestsInFlight)
	})
}

// Metrics HTTP 指标采集中间件
func Metrics(serviceName string) gin.HandlerFunc {
	initMetrics(serviceName)
	return func(c *gin.Context) {
		start := time.Now()
		httpRequestsInFlight.Inc()

		c.Next()

		httpRequestsInFlight.Dec()
		duration := float64(time.Since(start).Milliseconds())
		status := strconv.Itoa(c.Writer.Status())
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}
}

// MetricsHandler 返回 Prometheus /metrics 端点处理器
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
