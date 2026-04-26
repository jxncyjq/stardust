package metric

import "github.com/prometheus/client_golang/prometheus"

type CounterOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type CounterVec struct {
	*prometheus.CounterVec
}

func NewCounter(opts CounterOpts) *CounterVec {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, opts.Labels)
	prometheus.MustRegister(cv)
	return &CounterVec{cv}
}

func (c *CounterVec) Inc(labels ...string) {
	c.WithLabelValues(labels...).Inc()
}

func (c *CounterVec) Add(val float64, labels ...string) {
	c.WithLabelValues(labels...).Add(val)
}
