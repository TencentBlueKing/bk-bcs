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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientset "k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	heapster "k8s.io/heapster/metrics/api/v1/types"
	"k8s.io/klog"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1alpha1"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

const (
	// DefaultHeapsterNamespace DOTO
	DefaultHeapsterNamespace = "bcs-system"
	// DefaultHeapsterScheme DOTO
	DefaultHeapsterScheme = "http"
	// DefaultHeapsterService DOTO
	DefaultHeapsterService = "heapster"
	// DefaultHeapsterPort DOTO
	DefaultHeapsterPort         = "" // use the first exposed port on the service
	heapsterDefaultMetricWindow = time.Minute
)

var heapsterQueryStart = -5 * time.Minute

// HeapsterMetricsClient heapster metrics client
type HeapsterMetricsClient struct {
	services        v1core.ServiceInterface
	podsGetter      v1core.PodsGetter
	heapsterScheme  string
	heapsterService string
	heapsterPort    string
}

// NewHeapsterMetricsClient New Heapster Metrics Client
func NewHeapsterMetricsClient(client clientset.Interface, namespace, scheme, service, port string) MetricsClient {
	return &HeapsterMetricsClient{
		services:        client.CoreV1().Services(namespace),
		podsGetter:      client.CoreV1(),
		heapsterScheme:  scheme,
		heapsterService: service,
		heapsterPort:    port,
	}
}

// GetResourceMetric Get Resource Metric
func (h *HeapsterMetricsClient) GetResourceMetric(resource v1.ResourceName, namespace string, selector labels.Selector,
	container string) (PodMetricsInfo, time.Time, error) {
	metricPath := fmt.Sprintf("/apis/metrics/v1alpha1/namespaces/%s/pods", namespace)
	params := map[string]string{"labelSelector": selector.String()}

	resultRaw, err := h.services.
		ProxyGet(h.heapsterScheme, h.heapsterService, h.heapsterPort, metricPath, params).
		DoRaw()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get pod resource metrics: %v", err)
	}

	klog.V(8).Infof("Heapster metrics result: %s", string(resultRaw))

	metrics := metricsapi.PodMetricsList{}
	err = json.Unmarshal(resultRaw, &metrics)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to unmarshal heapster response: %v", err)
	}

	if len(metrics.Items) == 0 {
		return nil, time.Time{}, fmt.Errorf("no metrics returned from heapster")
	}

	res := make(PodMetricsInfo, len(metrics.Items))

	for _, m := range metrics.Items {
		podSum := int64(0)
		missing := len(m.Containers) == 0
		for _, c := range m.Containers {
			resValue, found := c.Usage[resource]
			if !found {
				missing = true
				klog.V(2).Infof("missing resource metric %v for container %s in pod %s/%s", resource, c.Name, namespace, m.Name)
				continue
			}
			podSum += resValue.MilliValue()
		}

		if !missing {
			res[m.Name] = PodMetric{
				Timestamp: m.Timestamp.Time,
				Window:    m.Window.Duration,
				Value:     podSum,
			}
		}
	}

	timestamp := metrics.Items[0].Timestamp.Time

	return res, timestamp, nil
}

// GetRawMetric Get Raw Metric
func (h *HeapsterMetricsClient) GetRawMetric(metricName string, namespace string, selector labels.Selector,
	metricSelector labels.Selector) (PodMetricsInfo, time.Time, error) {
	podList, err := h.podsGetter.Pods(namespace).List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get pod list while fetching metrics: %v", err)
	}

	if len(podList.Items) == 0 {
		return nil, time.Time{}, fmt.Errorf("no pods matched the provided selector")
	}

	podNames := make([]string, len(podList.Items))
	for i, pod := range podList.Items {
		podNames[i] = pod.Name
	}

	now := time.Now()

	startTime := now.Add(heapsterQueryStart)
	metricPath := fmt.Sprintf("/api/v1/model/namespaces/%s/pod-list/%s/metrics/%s",
		namespace,
		strings.Join(podNames, ","),
		metricName)

	resultRaw, err := h.services.
		ProxyGet(h.heapsterScheme, h.heapsterService, h.heapsterPort, metricPath, map[string]string{
			"start": startTime.Format(time.RFC3339)}).
		DoRaw()
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to get pod metrics: %v", err)
	}

	var metrics heapster.MetricResultList
	err = json.Unmarshal(resultRaw, &metrics)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("failed to unmarshal heapster response: %v", err)
	}

	klog.V(4).Infof("Heapster metrics result: %s", string(resultRaw))

	if len(metrics.Items) != len(podNames) {
		// if we get too many metrics or two few metrics, we have no way of knowing which metric goes to which pod
		// (note that Heapster returns *empty* metric items when a pod does not exist or have that metric, so this
		// does not cover the "missing metric entry" case)
		return nil, time.Time{}, fmt.Errorf("requested metrics for %v pods, got metrics for %v", len(podNames),
			len(metrics.Items))
	}

	var timestamp *time.Time
	res := make(PodMetricsInfo, len(metrics.Items))
	for i, podMetrics := range metrics.Items {
		val, podTimestamp, hadMetrics := collapseTimeSamples(podMetrics, time.Minute)
		if hadMetrics {
			res[podNames[i]] = PodMetric{
				Timestamp: podTimestamp,
				Window:    heapsterDefaultMetricWindow,
				Value:     val,
			}

			if timestamp == nil || podTimestamp.Before(*timestamp) {
				timestamp = &podTimestamp
			}
		}
	}

	if timestamp == nil {
		timestamp = &time.Time{}
	}

	return res, *timestamp, nil
}

// GetObjectMetric Get Object Metric
func (h *HeapsterMetricsClient) GetObjectMetric(
	metricName string,
	namespace string,
	objectRef *autoscaling.CrossVersionObjectReference,
	metricSelector labels.Selector) (int64, time.Time, error) {
	return 0, time.Time{}, fmt.Errorf("object metrics are not yet supported")
}

// GetExternalMetric Get External Metric
func (h *HeapsterMetricsClient) GetExternalMetric(metricName, namespace string, selector labels.Selector) ([]int64,
	time.Time, error) {
	return nil, time.Time{}, fmt.Errorf("external metrics aren't supported")
}

func collapseTimeSamples(metrics heapster.MetricResult, duration time.Duration) (int64, time.Time, bool) {
	floatSum := float64(0)
	intSum := int64(0)
	intSumCount := 0
	floatSumCount := 0

	var newest *heapster.MetricPoint // creation time of the newest sample for this pod
	for i, metricPoint := range metrics.Metrics {
		if newest == nil || newest.Timestamp.Before(metricPoint.Timestamp) {
			newest = &metrics.Metrics[i]
		}
	}
	if newest != nil {
		for _, metricPoint := range metrics.Metrics {
			if metricPoint.Timestamp.Add(duration).After(newest.Timestamp) {
				intSum += int64(metricPoint.Value)
				intSumCount++
				if metricPoint.FloatValue != nil {
					floatSum += *metricPoint.FloatValue
					floatSumCount++
				}
			}
		}

		if newest.FloatValue != nil {
			return int64(floatSum / float64(floatSumCount) * 1000), newest.Timestamp, true
		}
		return (intSum * 1000) / int64(intSumCount), newest.Timestamp, true
	}

	return 0, time.Time{}, false
}
