package metric

import "github.com/prometheus/client_golang/prometheus"

type HistogramOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Buckets   []float64
}

type HistogramVec struct {
	*prometheus.HistogramVec
}

func NewHistogram(opts HistogramOpts) *HistogramVec {
	buckets := opts.Buckets
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
		Buckets:   buckets,
	}, opts.Labels)
	prometheus.MustRegister(hv)
	return &HistogramVec{hv}
}

func (h *HistogramVec) Observe(val float64, labels ...string) {
	h.WithLabelValues(labels...).Observe(val)
}
