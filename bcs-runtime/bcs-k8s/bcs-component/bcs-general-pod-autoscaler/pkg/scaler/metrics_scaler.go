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

// Package scaler xxx
package scaler

import (
	"context"
	"fmt"
	"time"

	// "k8s.io/apimachinery/pkg/types"

	"github.com/pkg/errors"
	autoscalinginternal "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
)

// computeReplicasForMetrics computes the desired number of replicas for the metric specifications listed in the GPA,
// returning the maximum  of the computed replica counts, a description of the associated metric, and the statuses of
// all metrics computed.
// It may return both valid metricDesiredReplicas and an error,
// when some metrics still work and HPA should perform scaling based on them.
// If HPA cannot do anything due to error, it returns -1 in metricDesiredReplicas as a failure signal.
func (a *GeneralController) computeReplicasForMetrics(ctx context.Context, gpa *autoscaling.GeneralPodAutoscaler,
	scale *autoscalinginternal.Scale, metricSpecs []autoscaling.MetricSpec) (replicas int32, metric string,
	statuses []autoscaling.MetricStatus, timestamp time.Time, err error) {
	replicas = -1

	selector, err := a.validateAndParseSelector(gpa, scale.Status.Selector)
	if err != nil {
		return -1, "", nil, time.Time{}, err
	}

	specReplicas := scale.Spec.Replicas
	statusReplicas := scale.Status.Replicas
	statuses = make([]autoscaling.MetricStatus, len(metricSpecs))

	invalidMetricsCount := 0
	var invalidMetricError error
	var invalidMetricCondition autoscaling.GeneralPodAutoscalerCondition

	for i, metricSpec := range metricSpecs {
		startTime := time.Now()
		replicaCountProposal, metricNameProposal, timestampProposal, condition, err := a.computeReplicasForMetric(
			ctx, gpa, metricSpec, specReplicas, statusReplicas, selector, &statuses[i])
		if err != nil {
			metricsServer.RecordScalerExecDuration(gpa, getMetricName(metricSpec),
				"metric", "failure", time.Since(startTime))
			metricsServer.RecordScalerMetricExecDuration(gpa, getMetricName(metricSpec),
				"metric", "failure", time.Since(startTime))
			metricsServer.RecordGPAScalerError(gpa, "metric", getMetricName(metricSpec))
			if invalidMetricsCount <= 0 {
				invalidMetricCondition = condition
				invalidMetricError = err
			}
			invalidMetricsCount++
			continue
		}
		metricsServer.RecordScalerExecDuration(gpa, getMetricName(metricSpec), "metric", "success", time.Since(startTime))
		metricsServer.RecordScalerMetricExecDuration(gpa, getMetricName(metricSpec),
			"metric", "success", time.Since(startTime))
		if replicas == -1 || replicaCountProposal > replicas {
			timestamp = timestampProposal
			replicas = replicaCountProposal
			metric = metricNameProposal
		}
	}

	if invalidMetricError != nil {
		invalidMetricError = fmt.Errorf("invalid metrics (%v invalid out of %v), first error is: %v",
			invalidMetricsCount, len(metricSpecs), invalidMetricError)
	}

	// If all metrics are invalid or some are invalid and we would scale down,
	// return an error and set the condition of the hpa based on the first invalid metric.
	// Otherwise set the condition as scaling active as we're going to scale
	if invalidMetricsCount >= len(metricSpecs) || (invalidMetricsCount > 0 && replicas < specReplicas) {
		setCondition(gpa, invalidMetricCondition.Type, invalidMetricCondition.Status, invalidMetricCondition.Reason,
			invalidMetricCondition.Message)
		metricsServer.RecordGPAScalerDesiredReplicas(gpa, "metric", -1)
		return -1, "", statuses, time.Time{}, invalidMetricError
	}
	setCondition(gpa, autoscaling.ScalingActive, v1.ConditionTrue, "ValidMetricFound",
		"the GPA was able to successfully calculate a replica count from %s", metric)
	metricsServer.RecordGPAScalerDesiredReplicas(gpa, "metric", replicas)

	return replicas, metric, statuses, timestamp, nil
}

// validateAndParseSelector verifies that:
// - selector is not empty;
// - selector format is valid;
// - all pods by current selector are controlled by only one HPA.
// Returns an error if the check has failed or the parsed selector if succeeded.
// In case of an error the ScalingActive is set to false with the corresponding reason.
func (a *GeneralController) validateAndParseSelector(gpa *autoscaling.GeneralPodAutoscaler,
	selector string) (labels.Selector, error) {
	if selector == "" {
		errMsg := "selector is required"
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "SelectorRequired", errMsg)
		setCondition(gpa, autoscaling.ScalingActive, v1.ConditionFalse, "InvalidSelector",
			"the GPA target's scale is missing a selector")
		return nil, errors.New(errMsg)
	}

	parsedSelector, err := labels.Parse(selector)
	if err != nil {
		errMsg := fmt.Sprintf("couldn't convert selector into a corresponding internal selector object: %v", err)
		a.eventRecorder.Event(gpa, v1.EventTypeWarning, "InvalidSelector", errMsg)
		setCondition(gpa, autoscaling.ScalingActive, v1.ConditionFalse, "InvalidSelector", errMsg)
		return nil, errors.New(errMsg)
	}

	return parsedSelector, nil
}

// computeReplicasForMetric Computes the desired number of replicas for a specific gpa and metric specification,
// returning the metric status and a proposed condition to be set on the GPA object.
func (a *GeneralController) computeReplicasForMetric(ctx context.Context, gpa *autoscaling.GeneralPodAutoscaler,
	spec autoscaling.MetricSpec, specReplicas, statusReplicas int32,
	selector labels.Selector, status *autoscaling.MetricStatus) (replicaCountProposal int32,
	metricNameProposal string, timestampProposal time.Time, condition autoscaling.GeneralPodAutoscalerCondition,
	err error) {

	switch spec.Type {
	case autoscaling.ObjectMetricSourceType:
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Object.Metric.Selector) // nolint
		if err != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetObjectMetric", err)
			return 0, "", time.Time{}, condition,
				fmt.Errorf("failed to get object metric value: %v", err)
		}
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err =
			a.computeStatusForObjectMetric(specReplicas, statusReplicas, spec, gpa, selector, status, metricSelector)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get object metric value: %v", err)
		}
	case autoscaling.PodsMetricSourceType:
		metricSelector, err := metav1.LabelSelectorAsSelector(spec.Pods.Metric.Selector) // nolint
		if err != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetPodsMetric", err)
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get pods metric value: %v", err)
		}
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err =
			a.computeStatusForPodsMetric(specReplicas, spec, gpa, selector, status, metricSelector)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get pods metric value: %v", err)
		}
	case autoscaling.ResourceMetricSourceType:
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err =
			a.computeStatusForResourceMetric(ctx, specReplicas, spec, gpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s resource metric value: %v", spec.Resource.Name,
				err)
		}
	case autoscaling.ContainerResourceMetricSourceType:
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err =
			a.computeStatusForContainerResourceMetric(ctx, specReplicas, spec, gpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s container metric value: %v",
				spec.ContainerResource.Container, err)
		}
	case autoscaling.ExternalMetricSourceType:
		replicaCountProposal, timestampProposal, metricNameProposal, condition, err =
			a.computeStatusForExternalMetric(specReplicas, statusReplicas, spec, gpa, selector, status)
		if err != nil {
			return 0, "", time.Time{}, condition, fmt.Errorf("failed to get %s external metric value: %v",
				spec.External.Metric.Name, err)
		}
	default:
		err = fmt.Errorf("unknown metric source type %q", string(spec.Type))
		condition := a.getUnableComputeReplicaCountCondition(gpa, "InvalidMetricSourceType", err)
		return 0, "", time.Time{}, condition, err
	}
	return replicaCountProposal, metricNameProposal, timestampProposal,
		autoscaling.GeneralPodAutoscalerCondition{}, nil
}

// computeStatusForObjectMetric computes the desired number of replicas for
// the specified metric of type ObjectMetricSourceType.
func (a *GeneralController) computeStatusForObjectMetric(specReplicas, statusReplicas int32,
	metricSpec autoscaling.MetricSpec, gpa *autoscaling.GeneralPodAutoscaler, selector labels.Selector,
	status *autoscaling.MetricStatus, metricSelector labels.Selector) (replicas int32,
	timestamp time.Time, metricName string, condition autoscaling.GeneralPodAutoscalerCondition, err error) {
	if metricSpec.Object.Target.Type == autoscaling.ValueMetricType && metricSpec.Object.Target.Value != nil {
		replicaCountProposal, utilizationProposal, timestampProposal, err2 := a.replicaCalc.GetObjectMetricReplicas(
			specReplicas, metricSpec.Object.Target.Value.MilliValue(), metricSpec.Object.Metric.Name,
			gpa.Namespace, &metricSpec.Object.DescribedObject, selector, metricSelector)
		if err2 != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetObjectMetric", err2)
			return 0, timestampProposal, "", condition, err2
		}
		*status = autoscaling.MetricStatus{
			Type: autoscaling.ObjectMetricSourceType,
			Object: &autoscaling.ObjectMetricStatus{
				DescribedObject: metricSpec.Object.DescribedObject,
				Metric: autoscaling.MetricIdentifier{
					Name:     metricSpec.Object.Metric.Name,
					Selector: metricSpec.Object.Metric.Selector,
				},
				Current: autoscaling.MetricValueStatus{
					Value: resource.NewMilliQuantity(utilizationProposal, resource.DecimalSI),
				},
			},
		}
		metricsServer.RecordGPAScalerMetric(gpa, "metric", metricSpec.Object.Metric.Name,
			metricSpec.Object.Target.Value.Value(), status.Object.Current.Value.Value())
		return replicaCountProposal, timestampProposal,
			fmt.Sprintf("%s metric %s", metricSpec.Object.DescribedObject.Kind, metricSpec.Object.Metric.Name),
			autoscaling.GeneralPodAutoscalerCondition{}, nil
	} else if metricSpec.Object.Target.Type == autoscaling.AverageValueMetricType &&
		metricSpec.Object.Target.AverageValue != nil {
		replicaCountProposal, utilizationProposal, timestampProposal, err2 := a.replicaCalc.GetObjectPerPodMetricReplicas(
			statusReplicas, metricSpec.Object.Target.AverageValue.MilliValue(), metricSpec.Object.Metric.Name, gpa.Namespace,
			&metricSpec.Object.DescribedObject, metricSelector)
		if err2 != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetObjectMetric", err2)
			return 0, time.Time{}, "", condition,
				fmt.Errorf("failed to get %s object metric: %v", metricSpec.Object.Metric.Name, err2)
		}
		*status = autoscaling.MetricStatus{
			Type: autoscaling.ObjectMetricSourceType,
			Object: &autoscaling.ObjectMetricStatus{
				Metric: autoscaling.MetricIdentifier{
					Name:     metricSpec.Object.Metric.Name,
					Selector: metricSpec.Object.Metric.Selector,
				},
				Current: autoscaling.MetricValueStatus{
					AverageValue: resource.NewMilliQuantity(utilizationProposal, resource.DecimalSI),
				},
			},
		}
		metricsServer.RecordGPAScalerMetric(gpa, "metric", metricSpec.Object.Metric.Name,
			metricSpec.Object.Target.AverageValue.Value(), status.Object.Current.AverageValue.Value())
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)", metricSpec.Object.Metric.Name,
			metricSpec.Object.Metric.Selector), autoscaling.GeneralPodAutoscalerCondition{}, nil
	}
	errMsg := "invalid object metric source: neither a value target nor an average value target was set"
	err = errors.New(errMsg)
	condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetObjectMetric", err)
	return 0, time.Time{}, "", condition, err
}

// computeStatusForPodsMetric computes the desired number of replicas for the specified metric of
// type PodsMetricSourceType.
func (a *GeneralController) computeStatusForPodsMetric(currentReplicas int32, metricSpec autoscaling.MetricSpec,
	gpa *autoscaling.GeneralPodAutoscaler, selector labels.Selector, status *autoscaling.MetricStatus,
	metricSelector labels.Selector) (replicaCountProposal int32, timestampProposal time.Time,
	metricNameProposal string, condition autoscaling.GeneralPodAutoscalerCondition, err error) {
	replicaCountProposal, utilizationProposal, timestampProposal, err := a.replicaCalc.GetMetricReplicas(
		currentReplicas, metricSpec.Pods.Target.AverageValue.MilliValue(),
		metricSpec.Pods.Metric.Name, gpa.Namespace, selector, metricSelector)
	if err != nil {
		condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetPodsMetric", err)
		return 0, timestampProposal, "", condition, err
	}
	*status = autoscaling.MetricStatus{
		Type: autoscaling.PodsMetricSourceType,
		Pods: &autoscaling.PodsMetricStatus{
			Metric: autoscaling.MetricIdentifier{
				Name:     metricSpec.Pods.Metric.Name,
				Selector: metricSpec.Pods.Metric.Selector,
			},
			Current: autoscaling.MetricValueStatus{
				AverageValue: resource.NewMilliQuantity(utilizationProposal, resource.DecimalSI),
			},
		},
	}
	metricsServer.RecordGPAScalerMetric(gpa, "metric", metricSpec.Pods.Metric.Name,
		metricSpec.Pods.Target.AverageValue.Value(), status.Pods.Current.AverageValue.Value())
	return replicaCountProposal, timestampProposal, fmt.Sprintf("pods metric %s", metricSpec.Pods.Metric.Name),
		autoscaling.GeneralPodAutoscalerCondition{}, nil
}

// computeStatusForResourceMetricGeneric Computes the desired number of replicas for a specific gpa and metric specification,
// returning the metric status and a proposed condition to be set on the GPA object.
// nolint
func (a *GeneralController) computeStatusForResourceMetricGeneric(ctx context.Context, currentReplicas int32,
	target autoscaling.MetricTarget, resourceName v1.ResourceName,
	container string, selector labels.Selector, metricSpec autoscaling.MetricSpec,
	gpa *autoscaling.GeneralPodAutoscaler) (replicaCountProposal int32, metricStatus *autoscaling.MetricValueStatus,
	timestampProposal time.Time, metricNameProposal string, condition autoscaling.GeneralPodAutoscalerCondition,
	err error) {
	namespace := gpa.Namespace
	computeByLimits := isComputeByLimits(gpa)
	if target.AverageValue != nil {
		var rawProposal int64
		replicaCountProposal, rawProposal, timestampProposal, err = a.replicaCalc.GetRawResourceReplicas(
			ctx, currentReplicas, target.AverageValue.MilliValue(), resourceName,
			namespace, selector, container)
		if err != nil {
			return 0, nil, time.Time{}, "", condition,
				fmt.Errorf("failed to get %s utilization: %v", resourceName, err)
		}
		metricNameProposal = fmt.Sprintf("%s resource", resourceName.String())
		status := autoscaling.MetricValueStatus{
			AverageValue: resource.NewMilliQuantity(rawProposal, resource.DecimalSI),
		}
		if metricSpec.Type == autoscaling.ContainerResourceMetricSourceType {
			metricsServer.RecordGPAScalerMetric(gpa, "metric", string(metricSpec.ContainerResource.Name),
				metricSpec.ContainerResource.Target.AverageValue.Value(), status.AverageValue.Value())
		} else if metricSpec.Type == autoscaling.ResourceMetricSourceType {
			metricsServer.RecordGPAScalerMetric(gpa, "metric", string(metricSpec.Resource.Name),
				metricSpec.Resource.Target.AverageValue.Value(), status.AverageValue.Value())
		}

		return replicaCountProposal, &status, timestampProposal,
			metricNameProposal, autoscaling.GeneralPodAutoscalerCondition{}, nil
	}

	if target.AverageUtilization == nil {
		errMsg := "invalid resource metric source: neither a utilization target nor a value target was set"
		return 0, nil, time.Time{}, "", condition, errors.New(errMsg)
	}

	targetUtilization := *target.AverageUtilization
	replicaCountProposal, percentageProposal, rawProposal, timestampProposal, err := a.replicaCalc.GetResourceReplicas(
		ctx, currentReplicas, targetUtilization, resourceName, namespace,
		selector, container, computeByLimits)
	if err != nil {
		return 0, nil, time.Time{}, "", condition, fmt.Errorf("failed to get %s utilization: %v", resourceName, err)
	}

	computeResourceUtilizationRatioBy := "request"
	if computeByLimits {
		computeResourceUtilizationRatioBy = "limit"
	}

	status := autoscaling.MetricValueStatus{
		AverageUtilization: &percentageProposal,
		AverageValue:       resource.NewMilliQuantity(rawProposal, resource.DecimalSI),
	}

	if metricSpec.Type == autoscaling.ContainerResourceMetricSourceType {
		metricNameProposal = fmt.Sprintf("%s container resource utilization (percentage of %s)",
			resourceName, computeResourceUtilizationRatioBy)
		metricsServer.RecordGPAScalerMetric(gpa, "metric", string(metricSpec.ContainerResource.Name),
			int64(targetUtilization), int64(*status.AverageUtilization))
	} else if metricSpec.Type == autoscaling.ResourceMetricSourceType {
		metricNameProposal = fmt.Sprintf("%s resource utilization (percentage of %s)",
			resourceName, computeResourceUtilizationRatioBy)
		metricsServer.RecordGPAScalerMetric(gpa, "metric", string(metricSpec.Resource.Name), int64(targetUtilization),
			int64(*status.AverageUtilization))
	}

	return replicaCountProposal, &status, timestampProposal, metricNameProposal,
		autoscaling.GeneralPodAutoscalerCondition{}, nil
}

// computeStatusForResourceMetric computes the desired number of replicas for the specified metric of
// type ResourceMetricSourceType.
func (a *GeneralController) computeStatusForResourceMetric(ctx context.Context, currentReplicas int32,
	metricSpec autoscaling.MetricSpec, gpa *autoscaling.GeneralPodAutoscaler,
	selector labels.Selector, status *autoscaling.MetricStatus) (
	replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string,
	condition autoscaling.GeneralPodAutoscalerCondition, err error) {
	replicaCountProposal, metricValueStatus, timestampProposal, metricNameProposal, condition, err :=
		a.computeStatusForResourceMetricGeneric(ctx, currentReplicas, metricSpec.Resource.Target, metricSpec.Resource.Name,
			"", selector, metricSpec, gpa)
	if err != nil {
		condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetResourceMetric", err)
		return replicaCountProposal, timestampProposal, metricNameProposal, condition, err
	}
	*status = autoscaling.MetricStatus{
		Type: autoscaling.ResourceMetricSourceType,
		Resource: &autoscaling.ResourceMetricStatus{
			Name:    metricSpec.Resource.Name,
			Current: *metricValueStatus,
		},
	}

	return replicaCountProposal, timestampProposal, metricNameProposal,
		condition, nil
}

// computeStatusForContainerResourceMetric computes the desired number of replicas for the specified metric of
// type ResourceMetricSourceType.
// NOCC:tosa/fn_length(设计如此)
func (a *GeneralController) computeStatusForContainerResourceMetric(ctx context.Context, currentReplicas int32,
	metricSpec autoscaling.MetricSpec, gpa *autoscaling.GeneralPodAutoscaler,
	selector labels.Selector, status *autoscaling.MetricStatus) (replicaCountProposal int32,
	timestampProposal time.Time, metricNameProposal string, condition autoscaling.GeneralPodAutoscalerCondition,
	err error) {
	replicaCountProposal, metricValueStatus, timestampProposal, metricNameProposal, condition, err :=
		a.computeStatusForResourceMetricGeneric(ctx, currentReplicas, metricSpec.ContainerResource.Target,
			metricSpec.ContainerResource.Name, metricSpec.ContainerResource.Container,
			selector, metricSpec, gpa)
	if err != nil {
		condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetContainerResourceMetric", err)
		return replicaCountProposal, timestampProposal, metricNameProposal, condition, err
	}
	*status = autoscaling.MetricStatus{
		Type: autoscaling.ContainerResourceMetricSourceType,
		ContainerResource: &autoscaling.ContainerResourceMetricStatus{
			Name:      metricSpec.ContainerResource.Name,
			Container: metricSpec.ContainerResource.Container,
			Current:   *metricValueStatus,
		},
	}
	return replicaCountProposal, timestampProposal, metricNameProposal, condition, nil
}

// computeStatusForExternalMetric computes the desired number of replicas for the specified metric of
// type ExternalMetricSourceType.
func (a *GeneralController) computeStatusForExternalMetric(
	specReplicas, statusReplicas int32, metricSpec autoscaling.MetricSpec,
	gpa *autoscaling.GeneralPodAutoscaler, selector labels.Selector, status *autoscaling.MetricStatus) (
	replicaCountProposal int32, timestampProposal time.Time, metricNameProposal string,
	condition autoscaling.GeneralPodAutoscalerCondition, err error) {

	if metricSpec.External.Target.AverageValue != nil {
		replicaCountProposal, utilizationProposal, timestampProposal, err2 :=
			a.replicaCalc.GetExternalPerPodMetricReplicas(statusReplicas,
				metricSpec.External.Target.AverageValue.MilliValue(),
				metricSpec.External.Metric.Name, gpa.Namespace, metricSpec.External.Metric.Selector)
		if err2 != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetExternalMetric", err2)
			return 0, time.Time{}, "", condition,
				fmt.Errorf("failed to get %s external metric: %v", metricSpec.External.Metric.Name, err2)
		}
		*status = autoscaling.MetricStatus{
			Type: autoscaling.ExternalMetricSourceType,
			External: &autoscaling.ExternalMetricStatus{
				Metric: autoscaling.MetricIdentifier{
					Name:     metricSpec.External.Metric.Name,
					Selector: metricSpec.External.Metric.Selector,
				},
				Current: autoscaling.MetricValueStatus{
					AverageValue: resource.NewMilliQuantity(utilizationProposal, resource.DecimalSI),
				},
			},
		}
		metricsServer.RecordGPAScalerMetric(gpa, "metric", metricSpec.External.Metric.Name,
			metricSpec.External.Target.AverageValue.Value(), status.External.Current.AverageValue.Value())
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)",
				metricSpec.External.Metric.Name, metricSpec.External.Metric.Selector),
			autoscaling.GeneralPodAutoscalerCondition{}, nil
	}
	if metricSpec.External.Target.Value != nil {
		replicaCountProposal, utilizationProposal, timestampProposal, err2 :=
			a.replicaCalc.GetExternalMetricReplicas(specReplicas, metricSpec.External.Target.Value.MilliValue(),
				metricSpec.External.Metric.Name, gpa.Namespace, metricSpec.External.Metric.Selector, selector)
		if err2 != nil {
			condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetExternalMetric", err2)
			return 0, time.Time{}, "", condition,
				fmt.Errorf("failed to get external metric %s: %v", metricSpec.External.Metric.Name, err2)
		}
		*status = autoscaling.MetricStatus{
			Type: autoscaling.ExternalMetricSourceType,
			External: &autoscaling.ExternalMetricStatus{
				Metric: autoscaling.MetricIdentifier{
					Name:     metricSpec.External.Metric.Name,
					Selector: metricSpec.External.Metric.Selector,
				},
				Current: autoscaling.MetricValueStatus{
					Value: resource.NewMilliQuantity(utilizationProposal, resource.DecimalSI),
				},
			},
		}
		metricsServer.RecordGPAScalerMetric(gpa, "metric", metricSpec.External.Metric.Name,
			metricSpec.External.Target.Value.Value(), status.External.Current.Value.Value())
		return replicaCountProposal, timestampProposal, fmt.Sprintf("external metric %s(%+v)",
				metricSpec.External.Metric.Name, metricSpec.External.Metric.Selector),
			autoscaling.GeneralPodAutoscalerCondition{}, nil
	}
	errMsg := "invalid external metric source: neither a value target nor an average value target was set"
	err = errors.New(errMsg)
	condition = a.getUnableComputeReplicaCountCondition(gpa, "FailedGetExternalMetric", err)
	return 0, time.Time{}, "", condition, err
}
