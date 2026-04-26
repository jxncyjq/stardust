package metric

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCounter(t *testing.T) {
	counter := NewCounter(CounterOpts{
		Namespace: "stardust_test_c",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests",
		Labels:    []string{"method", "path", "code"},
	})
	assert.NotNil(t, counter)
	counter.Inc("GET", "/api/users", "200")
	counter.Add(5, "POST", "/api/users", "201")
}

func TestNewHistogram(t *testing.T) {
	histogram := NewHistogram(HistogramOpts{
		Namespace: "stardust_test_h",
		Subsystem: "http",
		Name:      "request_duration_ms",
		Help:      "HTTP request duration in milliseconds",
		Labels:    []string{"method", "path"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	})
	assert.NotNil(t, histogram)
	histogram.Observe(42.5, "GET", "/api/users")
}

func TestNewGauge(t *testing.T) {
	gauge := NewGauge(GaugeOpts{
		Namespace: "stardust_test_g",
		Subsystem: "server",
		Name:      "connections",
		Help:      "Current connections",
		Labels:    []string{"type"},
	})
	assert.NotNil(t, gauge)
	gauge.Set(10, "websocket")
	gauge.Inc("websocket")
	gauge.Add(-1, "websocket")
}
