package ai

import (
	"github.com/prometheus/client_golang/prometheus"
)

// NewLLMMetrics creates and registers metrics for LLM services
func NewLLMMetrics() *LLMMetrics {
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "llm_request_duration_seconds",
			Help:    "Duration of LLM API requests",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"provider", "status"},
	)

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "llm_requests_total",
			Help: "Total number of LLM API requests",
		},
		[]string{"provider", "status"},
	)

	// Register metrics with Prometheus
	prometheus.MustRegister(requestDuration, requestCounter)

	return &LLMMetrics{
		RequestDuration: requestDuration,
		RequestCounter:  requestCounter,
	}
}
