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

// Package estimator xxx
package estimator

import (
	"fmt"

	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

const (
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
	// OldBinpackingEstimatorName is the name of the older binpacking estimator.
	OldBinpackingEstimatorName = "oldbinpacking"
	// ResourceEstimatorName is the name of the resource estimator.
	ResourceEstimatorName = "clusterresource"

	oldBinPackingEstimatorDeprecationMessage = "old binpacking estimator is deprecated. " +
		"It will be removed in Cluster Autoscaler 1.15."
)

// ExtendedEstimatorBuilder creates a new estimator object.
type ExtendedEstimatorBuilder func(*simulator.PredicateChecker, map[string]*nodeinfo.NodeInfo) estimator.Estimator

// NewEstimatorBuilder creates a new estimator object from flag.
func NewEstimatorBuilder(name string,
	cpuRatio, memRatio, resourceRatio float64) (ExtendedEstimatorBuilder, error) {
	switch name {
	case BinpackingEstimatorName:
		return func(predicateChecker *simulator.PredicateChecker,
			nodeInfos map[string]*nodeinfo.NodeInfo) estimator.Estimator {
			return estimator.NewBinpackingNodeEstimator(predicateChecker)
		}, nil
	case ResourceEstimatorName:
		return func(predicateChecker *simulator.PredicateChecker,
			nodeInfos map[string]*nodeinfo.NodeInfo) estimator.Estimator {
			return NewClusterResourceEstimator(predicateChecker, nodeInfos,
				cpuRatio, memRatio, resourceRatio)
		}, nil
	// Deprecated.
	case OldBinpackingEstimatorName:
		klog.Warning(oldBinPackingEstimatorDeprecationMessage)
		return func(predicateChecker *simulator.PredicateChecker,
			nodeInfos map[string]*nodeinfo.NodeInfo) estimator.Estimator {
			return estimator.NewOldBinpackingNodeEstimator(predicateChecker)
		}, nil
	}
	return nil, fmt.Errorf("unknown estimator: %s", name)
}
