package metric

import "github.com/prometheus/client_golang/prometheus"

type GaugeOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
}

type GaugeVec struct {
	*prometheus.GaugeVec
}

func NewGauge(opts GaugeOpts) *GaugeVec {
	gv := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: opts.Namespace,
		Subsystem: opts.Subsystem,
		Name:      opts.Name,
		Help:      opts.Help,
	}, opts.Labels)
	prometheus.MustRegister(gv)
	return &GaugeVec{gv}
}

func (g *GaugeVec) Set(val float64, labels ...string) {
	g.WithLabelValues(labels...).Set(val)
}

func (g *GaugeVec) Inc(labels ...string) {
	g.WithLabelValues(labels...).Inc()
}

func (g *GaugeVec) Add(val float64, labels ...string) {
	g.WithLabelValues(labels...).Add(val)
}
