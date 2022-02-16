package tenants

import (
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func newHistogram() *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "tenant_translator_service_request_duration_seconds",
		Help: "Translator service request duration",
	}, []string{"operation", "result"})
}

type measuringOperationAwareDoerImpl struct {
	delegate operationAwareDoer
	observer *prometheus.HistogramVec
}

func (this *measuringOperationAwareDoerImpl) Do(req *http.Request, operation string) (resp *http.Response, err error) {
	t := time.Now()
	resp, err = this.delegate.Do(req, operation)
	d := time.Since(t)

	result := "error"
	if err == nil {
		result = fmt.Sprintf("%d", resp.StatusCode)
	}

	this.observer.WithLabelValues(operation, result).Observe(d.Seconds())

	return
}

func (this *measuringOperationAwareDoerImpl) Unwrap() utils.HttpRequestDoer {
	return this.delegate.Unwrap()
}
