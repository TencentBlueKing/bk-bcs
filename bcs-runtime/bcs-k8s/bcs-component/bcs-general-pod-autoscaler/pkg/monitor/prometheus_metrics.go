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

package monitor

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
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

var register sync.Once

func init() {
	register.Do(func() {
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
	})

}

// NewServer creates a new http serving instance of prometheus metrics
func (metricsServer PrometheusMetricServer) NewServer(address string, pattern string) {
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			klog.Fatalf("Unable to write to serve custom metrics: %v", err)
		}
	})
	klog.Infof("Starting metrics server at %v", address)
	http.Handle(pattern, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// initialize the total error metric
	_, errscaler := scalerErrorsTotal.GetMetricWith(prometheus.Labels{})
	if errscaler != nil {
		klog.Fatalf("Unable to initialize total error metrics as : %v", errscaler)
	}

	klog.Fatal(http.ListenAndServe(address, nil))
}

// RecordGPAScalerMetric create a measurement of the metric used by the GPA
// 每个模式的指标目标值和当前值
func (metricsServer PrometheusMetricServer) RecordGPAScalerMetric(gpa *autoscaling.GeneralPodAutoscaler,
	scaler string, metric string, targetValue int64, currentValue int64) {
	scalerTargetMetricsValue.With(getLabels(gpa, scaler, metric)).
		Set(float64(targetValue))
	scalerCurrentMetricsValue.With(getLabels(gpa, scaler, metric)).
		Set(float64(currentValue))
}

// RecordGPAScalerDesiredReplicas record desired replicas value computed by a scaling mode for GPA
// 每个模式的推荐副本数
func (metricsServer PrometheusMetricServer) RecordGPAScalerDesiredReplicas(gpa *autoscaling.GeneralPodAutoscaler,
	scaler string, replicas int32) {
	scalerDesiredReplicasValue.With(prometheus.Labels{"namespace": gpa.Namespace, "name": gpa.Name,
		"scaledObject": getTargetRefKey(gpa), "scaler": scaler}).Set(float64(replicas))
}

// RecordGPAReplicas record final replicas value for GPA
func (metricsServer PrometheusMetricServer) RecordGPAReplicas(gpa *autoscaling.GeneralPodAutoscaler,
	minReplicas, desiredReplicas int32) {

	gpaMinReplicasValue.With(getGPALabels(gpa)).Set(float64(minReplicas))
	gpaMaxReplicasValue.With(getGPALabels(gpa)).Set(float64(gpa.Spec.MaxReplicas))
	gpaDesiredReplicasValue.With(getGPALabels(gpa)).Set(float64(desiredReplicas))
	gpaCurrentReplicasValue.With(getGPALabels(gpa)).Set(float64(gpa.Status.CurrentReplicas))
}

// RecordScalerExecDuration records duration by seconds when executing scaler.
// In metric mode, it records duration of executing every metric.
func (metricsServer PrometheusMetricServer) RecordScalerExecDuration(gpa *autoscaling.GeneralPodAutoscaler, metric,
	scaler, status string, duration time.Duration) {
	scalerExecDuration.WithLabelValues(gpa.Namespace, gpa.Name, getTargetRefKey(gpa), metric, scaler, status).
		Observe(duration.Seconds())
}

// RecordScalerUpdateDuration records duration by seconds when updating a scale.
// histogram 类型
func (metricsServer PrometheusMetricServer) RecordScalerUpdateDuration(gpa *autoscaling.GeneralPodAutoscaler,
	status string, duration time.Duration) {
	scaleUpdateDuration.WithLabelValues(gpa.Namespace, gpa.Name, getTargetRefKey(gpa), status).Observe(duration.Seconds())
}

// RecordScalerMetricExecDuration records duration by second when executing metric
func (metricsServer PrometheusMetricServer) RecordScalerMetricExecDuration(gpa *autoscaling.GeneralPodAutoscaler,
	metric, scaler, status string, duration time.Duration) {
	scalerMetricExecDuration.WithLabelValues(gpa.Namespace, gpa.Name, getTargetRefKey(gpa), metric, scaler, status).
		Set(duration.Seconds())
}

// RecordScalerReplicasUpdateDuration records duration by seconds when updating a scale.
// gauge 类型
func (metricsServer PrometheusMetricServer) RecordScalerReplicasUpdateDuration(gpa *autoscaling.GeneralPodAutoscaler,
	status string, duration time.Duration) {
	scalerReplicasUpdateDuration.WithLabelValues(gpa.Namespace, gpa.Name, getTargetRefKey(gpa), status).
		Set(duration.Seconds())
}

// RecordGPAScalerError counts the number of errors occurred in trying get an external metric used by the GPA
func (metricsServer PrometheusMetricServer) RecordGPAScalerError(gpa *autoscaling.GeneralPodAutoscaler,
	scaler string, metric string) {
	// 模式错误
	scaledObjectErrors.With(getGPALabels(gpa)).Inc()
	// gpa 错误
	scalerErrors.With(getLabels(gpa, scaler, metric)).Inc()
	// 总错误
	scalerErrorsTotal.With(prometheus.Labels{}).Inc()
	return

}

func getLabels(gpa *autoscaling.GeneralPodAutoscaler, scaler string, metric string) prometheus.Labels {
	return prometheus.Labels{"namespace": gpa.Namespace, "name": gpa.Name, "scaledObject": getTargetRefKey(gpa),
		"scaler": scaler, "metric": metric}
}

func getGPALabels(gpa *autoscaling.GeneralPodAutoscaler) prometheus.Labels {
	return prometheus.Labels{"namespace": gpa.Namespace, "name": gpa.Name,
		"scaledObject": getTargetRefKey(gpa)}
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

func getTargetRefKey(gpa *autoscaling.GeneralPodAutoscaler) string {
	return fmt.Sprintf("%s/%s", gpa.Spec.ScaleTargetRef.Kind, gpa.Spec.ScaleTargetRef.Name)
}
