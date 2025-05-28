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
	"fmt"
	"time"

	pkgerrors "github.com/pkg/errors"
	autoscalinginternal "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	autoscaling "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/apis/autoscaling/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-general-pod-autoscaler/pkg/scalercore"
)

// computeReplicasForSimple computes the desired number of replicas for the metric specifications listed in the GPA,
// returning the maximum  of the computed replica counts, a description of the associated metric, and the statuses of
// all metrics computed.
// nolint
func (a *GeneralController) computeReplicasForSimple(gpa *autoscaling.GeneralPodAutoscaler,
	scale *autoscalinginternal.Scale, scaler scalercore.Scaler) (replicas int32, metric string,
	statuses []autoscaling.MetricStatus,
	timestamp time.Time, err error) {
	currentReplicas := scale.Spec.Replicas

	replicaCountProposal, modeNameProposal, err := computeDesiredSize(gpa, scaler, currentReplicas)
	if err != nil {
		setCondition(gpa, autoscaling.ScalingActive, v1.ConditionFalse, fmt.Sprintf("%v failed", modeNameProposal),
			fmt.Sprintf("%v failed: %v", modeNameProposal, err))
		return -1, "", statuses, time.Time{}, fmt.Errorf("invalid mode %v, first error is: %v", modeNameProposal, err)
	}

	replicas = replicaCountProposal
	metric = modeNameProposal
	setCondition(gpa, autoscaling.ScalingActive, v1.ConditionTrue, "ValidMetricFound",
		"the GPA was able to successfully calculate a replica count from %s", metric)
	timestamp = time.Now()
	return replicas, metric, statuses, timestamp, nil
}

// computeDesiredSize computes the new desired size of the given fleet
func computeDesiredSize(gpa *autoscaling.GeneralPodAutoscaler,
	scaler scalercore.Scaler, currentReplicas int32) (int32, string, error) {
	var (
		replicas int32
		err      error
	)
	replicas, err = scaler.GetReplicas(gpa, currentReplicas)
	if err != nil {
		klog.Error(err)
		err = pkgerrors.Wrap(err,
			fmt.Sprintf("GPA: %v get replicas error when call %v", gpa.Name, scaler.ScalerName()))
		return -1, scaler.ScalerName(), err
	}

	return replicas, scaler.ScalerName(), err
}
