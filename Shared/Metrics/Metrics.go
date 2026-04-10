package Metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var GRPCRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "grpc_requests_total",
	Help: "Total grpc request by method and status code",
},
	[]string{"method", "status_code"},
)
