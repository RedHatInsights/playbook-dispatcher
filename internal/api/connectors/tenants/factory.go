package tenants

import (
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"sort"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const defaultTimeout = 10

func NewTenantIDTranslatorClient(serviceHost string, options ...TenantIDTranslatorOption) TenantIDTranslator {
	sort.SliceStable(options, func(i, j int) bool {
		return options[i].priority() < options[j].priority()
	})

	instance := &translatorClientImpl{
		serviceHost: serviceHost,
		client: &operationAwareDoerImpl{
			client: &http.Client{
				Timeout: defaultTimeout * time.Second,
			},
		},
	}

	for _, option := range options {
		option.apply(instance)
	}

	return instance
}

//
// Configurable timeout
//
func WithTimeout(timeout time.Duration) TenantIDTranslatorOption {
	return &withTimeout{
		timeout: timeout,
	}
}

type withTimeout struct {
	timeout time.Duration
}

func (this *withTimeout) apply(impl *translatorClientImpl) {
	impl.client = &operationAwareDoerImpl{
		client: &http.Client{
			Timeout: this.timeout,
		},
	}
}

func (*withTimeout) priority() int {
	return 10
}

//
// Wrap or replace the HTTP client
//
func WithDoer(doer utils.HttpRequestDoer) TenantIDTranslatorOption {
	return &withDoer{
		fn: func(ignored utils.HttpRequestDoer) utils.HttpRequestDoer {
			return doer
		},
	}
}

func WithDoerWrapper(fn func(utils.HttpRequestDoer) utils.HttpRequestDoer) TenantIDTranslatorOption {
	return &withDoer{
		fn: fn,
	}
}

type withDoer struct {
	fn func(utils.HttpRequestDoer) utils.HttpRequestDoer
}

func (this *withDoer) apply(impl *translatorClientImpl) {
	impl.client = &operationAwareDoerImpl{
		client: this.fn(impl.client.Unwrap()),
	}
}

func (*withDoer) priority() int {
	return 20
}

//
// Metrics
//
func WithMetrics() TenantIDTranslatorOption {
	return &withMetrics{
		registerer: prometheus.DefaultRegisterer,
	}
}

func WithMetricsWithCustomRegisterer(registerer prometheus.Registerer) TenantIDTranslatorOption {
	return &withMetrics{
		registerer: registerer,
	}
}

type withMetrics struct {
	registerer prometheus.Registerer
}

func (this *withMetrics) apply(impl *translatorClientImpl) {
	observer := newHistogram()
	this.registerer.MustRegister(observer)

	impl.client = &measuringOperationAwareDoerImpl{
		delegate: impl.client,
		observer: observer,
	}
}

func (*withMetrics) priority() int {
	return 30
}
