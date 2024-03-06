/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricLabels      = []string{"namespace", "name", "metric", "scaledObject", "scaler"}
	scalerErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "errors_total",
			Help:      "Total number of errors for all scalers",
		},
		[]string{},
	)
	scalerTargetMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "target_metrics_value",
			Help:      "Target Metric Value used for GPA",
		},
		metricLabels,
	)
	scalerCurrentMetricsValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "current_metrics_value",
			Help:      "Current Metric Value used for GPA",
		},
		metricLabels,
	)
	scalerDesiredReplicasValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "desired_replicas_value",
			Help:      "Desired Replicas Value computed by a scaling mode for GPA",
		},
		[]string{"namespace", "name", "scaledObject", "scaler"},
	)
	scalerErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "errors",
			Help:      "Number of scaler errors",
		},
		metricLabels,
	)
	scalerExecDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaler",
			Name:      "exec_duration",
			Help:      "Duration(seconds) of executing scaler",
		},
		[]string{"namespace", "name", "scaledObject", "metric", "scaler", "status"},
	)
	scaleUpdateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "update_duration",
			Help:      "Duration(seconds) of updating scale",
		},
		[]string{"namespace", "name", "scaledObject", "status"},
	)
	scalerMetricExecDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "exec_duration",
			Help:      "Duration(seconds) of executing metric in Gauge",
		},
		[]string{"namespace", "name", "scaledObject", "metric", "scaler", "status"},
	)
	scalerReplicasUpdateDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "replicas_update_duration",
			Help:      "Duration(seconds) of updating replicas in Gauge",
		},
		[]string{"namespace", "name", "scaledObject", "status"},
	)
	scaledObjectErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "scaled_object",
			Name:      "errors",
			Help:      "Number of scaled object errors",
		},
		[]string{"namespace", "name", "scaledObject"},
	)
	gpaDesiredReplicasValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "desired_replicas_value",
			Help:      "Desired Replicas Value of a GPA",
		},
		[]string{"namespace", "name", "scaledObject"},
	)
	gpaCurrentReplicasValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "current_replicas_value",
			Help:      "Current Replicas Value of a GPA",
		},
		[]string{"namespace", "name", "scaledObject"},
	)
	gpaMinReplicasValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "min_replicas_value",
			Help:      "Min Replicas Value of a GPA",
		},
		[]string{"namespace", "name", "scaledObject"},
	)
	gpaMaxReplicasValue = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "keda_metrics_adapter",
			Subsystem: "gpa",
			Name:      "max_replicas_value",
			Help:      "Max Replicas Value of a GPA",
		},
		[]string{"namespace", "name", "scaledObject"},
	)
)

// PrometheusMetricServer the type of MetricsServer
type PrometheusMetricServer struct{}

var registry *prometheus.Registry

func init() {
	registry = prometheus.NewRegistry()
	registry.MustRegister(scalerErrorsTotal)
	registry.MustRegister(scalerTargetMetricsValue)
	registry.MustRegister(scalerCurrentMetricsValue)
	registry.MustRegister(scalerDesiredReplicasValue)
	registry.MustRegister(scalerErrors)
	registry.MustRegister(scaledObjectErrors)
	registry.MustRegister(gpaDesiredReplicasValue)
	registry.MustRegister(gpaCurrentReplicasValue)
	registry.MustRegister(gpaMinReplicasValue)
	registry.MustRegister(gpaMaxReplicasValue)
	registry.MustRegister(scalerExecDuration)
	registry.MustRegister(scaleUpdateDuration)
	registry.MustRegister(scalerMetricExecDuration)
	registry.MustRegister(scalerReplicasUpdateDuration)
}

// NewServer creates a new http serving instance of prometheus metrics
func (metricsServer PrometheusMetricServer) NewServer(address string, pattern string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatalf("Unable to write to serve custom metrics: %v", err)
		}
	})
	log.Printf("Starting metrics server at %v", address)
	http.Handle(pattern, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// initialize the total error metric
	_, errscaler := scalerErrorsTotal.GetMetricWith(prometheus.Labels{})
	if errscaler != nil {
		log.Fatalf("Unable to initialize total error metrics as : %v", errscaler)
	}

	log.Fatal(http.ListenAndServe(address, nil))
}

// RecordGPAScalerMetric create a measurement of the external metric used by the GPA
func (metricsServer PrometheusMetricServer) RecordGPAScalerMetric(namespace string, name string, scaledObject string,
	scaler string, metric string, targetValue int64, currentValue int64) {
	scalerTargetMetricsValue.With(getLabels(namespace, name, scaledObject, scaler, metric)).Set(float64(targetValue))
	scalerCurrentMetricsValue.With(getLabels(namespace, name, scaledObject, scaler, metric)).Set(float64(currentValue))
}

// RecordGPAScalerDesiredReplicas record desired replicas value computed by a scaling mode for GPA
func (metricsServer PrometheusMetricServer) RecordGPAScalerDesiredReplicas(namespace string, name string,
	scaledObject string, scaler string, replicas int32) {
	scalerDesiredReplicasValue.With(prometheus.Labels{"namespace": namespace, "name": name,
		"scaledObject": scaledObject, "scaler": scaler}).Set(float64(replicas))
}

// RecordGPAReplicas record final replicas value for GPA
func (metricsServer PrometheusMetricServer) RecordGPAReplicas(namespace string, name string, scaledObject string,
	minReplicas, maxReplicas, desiredReplicas, currentReplicas int32) {
	gpaMinReplicasValue.With(prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject}).
		Set(float64(minReplicas))
	gpaMaxReplicasValue.With(prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject}).
		Set(float64(maxReplicas))
	gpaDesiredReplicasValue.With(prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject}).
		Set(float64(desiredReplicas))
	gpaCurrentReplicasValue.With(prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject}).
		Set(float64(currentReplicas))
}

// RecordScalerExecDuration records duration by seconds when executing scaler.
// In metric mode, it records duration of executing every metric.
func (metricsServer PrometheusMetricServer) RecordScalerExecDuration(namespace, name, scaledObject, metric, scaler,
	status string, duration time.Duration) {
	scalerExecDuration.WithLabelValues(namespace, name, scaledObject, metric, scaler, status).Observe(duration.Seconds())
}

// RecordScalerUpdateDuration records duration by seconds when updating a scale.
func (metricsServer PrometheusMetricServer) RecordScalerUpdateDuration(namespace, name, scaledObject,
	status string, duration time.Duration) {
	scaleUpdateDuration.WithLabelValues(namespace, name, scaledObject,
		status).Observe(duration.Seconds())
}

// RecordScalerMetricExecDuration records duration by second when executing metric
func (metricsServer PrometheusMetricServer) RecordScalerMetricExecDuration(namespace, name, scaledObject, metric,
	scaler, status string, duration time.Duration) {
	scalerMetricExecDuration.WithLabelValues(namespace, name, scaledObject, metric, scaler, status).Set(duration.Seconds())
}

// RecordScalerReplicasUpdateDuration records duration by seconds when updating a scale.
func (metricsServer PrometheusMetricServer) RecordScalerReplicasUpdateDuration(namespace, name, scaledObject,
	status string, duration time.Duration) {
	scalerReplicasUpdateDuration.WithLabelValues(namespace, name, scaledObject, status).Set(duration.Seconds())
}

// RecordGPAScalerError counts the number of errors occurred in trying get an external metric used by the GPA
func (metricsServer PrometheusMetricServer) RecordGPAScalerError(namespace string, name string, scaledObject string,
	scaler string, metric string, isErr bool) {
	if isErr {
		scalerErrors.With(getLabels(namespace, name, scaledObject, scaler, metric)).Inc()
		// scaledObjectErrors.With(prometheus.Labels{"namespace": namespace, "scaledObject": scaledObject}).Inc()
		metricsServer.RecordScalerObjectError(namespace, name, scaledObject, isErr)
		scalerErrorsTotal.With(prometheus.Labels{}).Inc()
		return
	}
	// initialize metric with 0 if not already set
	_, errscaler := scalerErrors.GetMetricWith(getLabels(namespace, name, scaledObject, scaler, metric))
	_, errscaledobject := scaledObjectErrors.GetMetricWith(
		prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject})
	if errscaler != nil {
		log.Fatalf("Unable to write to serve custom metrics: %v", errscaler)
		return
	}
	if errscaledobject != nil {
		log.Fatalf("Unable to write to serve custom metrics: %v", errscaledobject)
		return
	}
}

// RecordScalerObjectError counts the number of errors with the scaled object
func (metricsServer PrometheusMetricServer) RecordScalerObjectError(namespace string, name string,
	scaledObject string, isErr bool) {
	labels := prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject}
	if isErr {
		scaledObjectErrors.With(labels).Inc()
		return
	}
	// initialize metric with 0 if not already set
	_, errscaledobject := scaledObjectErrors.GetMetricWith(labels)
	if errscaledobject != nil {
		log.Fatalf("Unable to write to serve custom metrics: %v", errscaledobject)
		return
	}
}

func getLabels(namespace string, name string, scaledObject string, scaler string, metric string) prometheus.Labels {
	return prometheus.Labels{"namespace": namespace, "name": name, "scaledObject": scaledObject,
		"scaler": scaler, "metric": metric}
}

// ResetScalerMetrics reset metrics when delete gpa object
func (metricsServer PrometheusMetricServer) ResetScalerMetrics(namespace, name string) {
	labels := prometheus.Labels{"namespace": namespace, "name": name}

	scalerTargetMetricsValue.DeletePartialMatch(labels)
	scalerCurrentMetricsValue.DeletePartialMatch(labels)
	scalerDesiredReplicasValue.DeletePartialMatch(labels)
	scalerMetricExecDuration.DeletePartialMatch(labels)
	scalerReplicasUpdateDuration.DeletePartialMatch(labels)
	gpaDesiredReplicasValue.DeletePartialMatch(labels)
	gpaCurrentReplicasValue.DeletePartialMatch(labels)
	gpaMinReplicasValue.DeletePartialMatch(labels)
	gpaMaxReplicasValue.DeletePartialMatch(labels)
}
