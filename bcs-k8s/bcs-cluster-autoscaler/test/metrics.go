package main

import (
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client-go metrics registration

	"github.com/prometheus/client_golang/prometheus"
)

const (
	caNamespace = "cluster_autoscaler_e2e"
)

var (
	failedScaleUpCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_up_count",
			Help:      "failed scale up count.",
		},
	)

	scaleUpCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scale_up_count",
			Help:      "scale up count.",
		},
	)

	failedScaleDownCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "failed_scale_down_count",
			Help:      "failed scale down count.",
		},
	)

	scaleDownCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: caNamespace,
			Name:      "scale_down_count",
			Help:      "scale down count.",
		},
	)

	scaleUpSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: caNamespace,
			Name:      "scale_up_seconds",
			Buckets:   []float64{60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		},
	)

	scaleDownSeconds = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: caNamespace,
			Name:      "scale_down_seconds",
			Buckets:   []float64{60, 120, 180, 240, 300, 360, 420, 480, 540, 600},
		},
	)

	scaleUpSuccessRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_up_success_rate",
		},
	)
	scaleDownSuccessRate = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: caNamespace,
			Name:      "scale_down_success_rate",
		},
	)
)

// registerAll registers all metrics.
func registerAll() {
	prometheus.MustRegister(failedScaleUpCount)
	prometheus.MustRegister(failedScaleDownCount)
	prometheus.MustRegister(scaleUpCount)
	prometheus.MustRegister(scaleDownCount)
	prometheus.MustRegister(scaleUpSeconds)
	prometheus.MustRegister(scaleDownSeconds)
	prometheus.MustRegister(scaleUpSuccessRate)
	prometheus.MustRegister(scaleDownSuccessRate)
}
