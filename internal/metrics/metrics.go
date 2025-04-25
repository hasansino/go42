package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RpsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "application_requests_total"},
		[]string{"status"},
	)
)
