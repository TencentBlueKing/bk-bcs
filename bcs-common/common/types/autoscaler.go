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
 *
 */

package types

import (
	"fmt"
	"time"
)

type AutoscalerTargetRefKind = string

const (
	AutoscalerTargetRefDeployment  AutoscalerTargetRefKind = "deployment"
	AutoscalerTargetRefApplication AutoscalerTargetRefKind = "application"
)

type BcsAutoscaler struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata"`
	//specification
	Spec *BcsAutoscalerSpec
	//status
	Status *BcsAutoscalerStatus
}

type BcsAutoscalerSpec struct {
	//target application
	ScaleTargetRef *TargetRef
	//min instance
	MinInstance uint
	//max instance
	MaxInstance uint
	//autoscale target metric infos
	MetricsTarget []*AutoscalerMetricTarget `json:"metrics"`
}

type AutoscalerMetricTarget struct {
	Type MetricSourceType
	//metric name
	Name string
	// description info
	Description string
	//autoscale metric target value
	Target *AutoscalerMetricValue
}

type MetricSourceType string

const (
	// ResourceMetricSourceType is a resource metric known to mesos, as
	// specified in limits, describing each taskgroup in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// mesos, and have special scaling options on top of those available
	// to normal per-taskgroup metrics (the "taskgroups" source).
	//target kind = AutoscalerMetricAverageUtilization
	ResourceMetricSourceType MetricSourceType = "Resource"
	// TaskgroupsMetricSourceType is a metric describing each taskgroup in the current scale
	// target (for example, transactions-processed-per-second).  The values
	// will be averaged together before being compared to the target value.
	//target kind = AutoscalerMetricTargetAverageValue
	TaskgroupsMetricSourceType MetricSourceType = "Taskgroup"
	// ExternalMetricSourceType is a global metric that is not associated
	// with any mesos object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	//target kind = AutoscalerMetricTargetValue
	ExternalMetricSourceType MetricSourceType = "External"
)

type AutoscalerMetricValue struct {
	Type AutoscalerMetricKind

	//kind = AutoscalerMetricAverageUtilization
	AverageUtilization float32 `json:"averageUtilization,omitempty"`
	//kind = AutoscalerMetricTargetAverageValue
	AverageValue float32 `json:"targetAverageValue,omitempty"`
	//kind = AutoscalerMetricTargetValue
	Value float32 `json:"targetValue,omitempty"`
}

type AutoscalerMetricKind string

const (
	//describing each pod in the current scale target (e.g. CPU or memory), The values will be averaged
	//together before being compared to the target
	AutoscalerMetricAverageUtilization AutoscalerMetricKind = "AverageUtilization"
	//The values will be averaged together pods before being compared to the target value
	AutoscalerMetricTargetAverageValue AutoscalerMetricKind = "AverageValue"
	//value is the target value of the metric
	AutoscalerMetricTargetValue AutoscalerMetricKind = "Value"
)

type BcsAutoscalerStatus struct {
	//autoscale application number, default: 0
	ScaleNumber uint
	//last time scale application
	LastScaleTime time.Time

	LastScaleOPeratorType AutoscalerOperatorType
	//current instance numbers
	CurrentInstance uint
	//desired instance numbers
	DesiredInstance uint
	//status
	CurrentMetrics []*AutoscalerMetricCurrent
	//target ref status
	TargetRefStatus string
}

const (
	TargetRefStatusNone = "none"
)

func (scaler *BcsAutoscaler) GetSpecifyCurrentMetrics(soucreType MetricSourceType, name string) (*AutoscalerMetricCurrent, error) {
	if scaler.Status == nil || len(scaler.Status.CurrentMetrics) == 0 {
		return nil, fmt.Errorf("Metrics type %s name %s not found", soucreType, name)
	}

	for _, current := range scaler.Status.CurrentMetrics {
		if current.Type == soucreType && current.Name == name {
			return current, nil
		}
	}

	return nil, fmt.Errorf("Metrics type %s name %s not found", soucreType, name)
}

type AutoscalerOperatorType string

const (
	AutoscalerOperatorNone      AutoscalerOperatorType = "none"
	AutoscalerOperatorScaleUp   AutoscalerOperatorType = "scaleUp"
	AutoscalerOperatorScaleDown AutoscalerOperatorType = "scaleDown"
)

type AutoscalerMetricCurrent struct {
	Type MetricSourceType
	//metric name
	Name string
	// description info
	Description string
	//autoscaler current metrics
	Current *AutoscalerMetricValue
	//timestamp
	Timestamp time.Time
}

func (scaler *BcsAutoscaler) GetUuid() string {
	uuid := fmt.Sprintf("%s_%s_%d", scaler.NameSpace, scaler.Name, scaler.CreationTimestamp.Unix())
	return uuid
}

func (scaler *BcsAutoscaler) InitAutoscalerStatus() {
	//scaler.CreationTimestamp = time.Now()
	scaler.Spec.ScaleTargetRef.Namespace = scaler.ObjectMeta.NameSpace
	currents := make([]*AutoscalerMetricCurrent, 0)
	for _, target := range scaler.Spec.MetricsTarget {
		current := &AutoscalerMetricCurrent{
			Type:        target.Type,
			Name:        target.Name,
			Description: target.Description,
			Current: &AutoscalerMetricValue{
				Type: target.Target.Type,
			},
		}

		currents = append(currents, current)
	}

	status := &BcsAutoscalerStatus{
		CurrentMetrics:        currents,
		LastScaleOPeratorType: AutoscalerOperatorNone,
		TargetRefStatus:       TargetRefStatusNone,
	}

	scaler.Status = status
}
