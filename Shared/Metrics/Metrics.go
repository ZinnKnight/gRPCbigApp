package Metrics

import (
	"gRPCbigapp/Shared/Metrics/MetricsPort"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ MetricsPort.MetricsRecord = (*PrometheusRecord)(nil)

type PrometheusRecord struct {
	registry *prometheus.Registry
	reqTotal *prometheus.CounterVec
}

func NewPrometheusRecord() *PrometheusRecord {
	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	reqTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_request_total",
		Help: "Total grpc request, including status codes and methods",
	}, []string{"method", "status_codes"})
	reg.MustRegister(reqTotal)

	return &PrometheusRecord{
		registry: reg,
		reqTotal: reqTotal,
	}
}

func (prom *PrometheusRecord) IncRequest(method, statusCode string) {
	prom.reqTotal.WithLabelValues(method, statusCode).Inc()
}

func (prom *PrometheusRecord) Registry() http.Handler {
	return promhttp.HandlerFor(prom.registry, promhttp.HandlerOpts{})
}
