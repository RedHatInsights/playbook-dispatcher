package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var baseHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "client_request_duration_seconds",
	Help: "Time spent talking to a service",
}, []string{"component", "operation", "result"})

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewMeasuredHttpRequestDoer(delegate HttpRequestDoer, component, operation string) HttpRequestDoer {
	return &measuredHttpRequestDoer{
		delegate: delegate,
		observer: baseHistogram.MustCurryWith(prometheus.Labels{"component": component, "operation": operation}),
	}
}

type measuredHttpRequestDoer struct {
	delegate HttpRequestDoer
	observer prometheus.ObserverVec
}

func (this *measuredHttpRequestDoer) Do(req *http.Request) (resp *http.Response, err error) {
	t := time.Now()
	resp, err = this.delegate.Do(req)
	d := time.Since(t)

	result := "error"
	if err == nil {
		result = fmt.Sprintf("%d", resp.StatusCode)
	}

	this.observer.WithLabelValues(result).Observe(d.Seconds())
	return
}
