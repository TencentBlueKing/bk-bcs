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
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

// PodMetric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type PodMetric struct {
	Timestamp time.Time
	Window    time.Duration
	Value     int64
}

// PodMetricsInfo contains pod metrics as a map from pod names to PodMetricsInfo
type PodMetricsInfo map[string]PodMetric

// MetricsClient knows how to query a remote interface to retrieve container-level
// resource metrics as well as pod-level arbitrary metrics
// nolint
type MetricsClient interface {
	// GetResourceMetric gets the given resource metric (and an associated oldest timestamp)
	// for all pods matching the specified selector in the given namespace
	GetResourceMetric(resource v1.ResourceName, namespace string, selector labels.Selector,
		container string) (PodMetricsInfo, time.Time, error)

	// GetRawMetric gets the given metric (and an associated oldest timestamp)
	// for all pods matching the specified selector in the given namespace
	GetRawMetric(metricName string, namespace string, selector labels.Selector,
		metricSelector labels.Selector) (PodMetricsInfo, time.Time, error)

	// GetObjectMetric gets the given metric (and an associated timestamp) for the given
	// object in the given namespace
	GetObjectMetric(metricName string, namespace string, objectRef *autoscaling.CrossVersionObjectReference,
		metricSelector labels.Selector) (int64, time.Time, error)

	// GetExternalMetric gets all the values of a given external metric
	// that match the specified selector.
	GetExternalMetric(metricName string, namespace string, selector labels.Selector) ([]int64, time.Time, error)
}
