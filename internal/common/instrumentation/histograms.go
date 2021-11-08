package instrumentation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var outboundHTTPDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "outbound_http_duration_seconds",
	Help: "Time spent waiting for an HTTP response",
	Buckets: []float64{
		0.0005,
		0.001, // 1ms
		0.002,
		0.005,
		0.01, // 10ms
		0.02,
		0.05,
		0.1, // 100 ms
		0.2,
		0.5,
		1.0, // 1s
		2.0,
		5.0,
		10.0, // 10s
		15.0,
		20.0,
		30.0,
	},
}, []string{"service"})

func OutboundHTTPDurationTimerFactory(service string) func() *prometheus.Timer {
	return func() *prometheus.Timer {
		return prometheus.NewTimer(outboundHTTPDurationHistogram.WithLabelValues(service))
	}
}
