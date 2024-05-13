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
)

const (
	// BinpackingEstimatorName is the name of binpacking estimator.
	BinpackingEstimatorName = "binpacking"
	// ResourceEstimatorName is the name of the resource estimator.
	ResourceEstimatorName = "clusterresource"
)

// AvailableEstimators is a list of available estimators.
var AvailableEstimators = []string{BinpackingEstimatorName, ResourceEstimatorName}

// ExtendedEstimatorBuilder creates a new estimator object.
type ExtendedEstimatorBuilder func(simulator.PredicateChecker, simulator.ClusterSnapshot) estimator.Estimator

// NewEstimatorBuilder creates a new estimator object from flag.
func NewEstimatorBuilder(name string,
	cpuRatio, memRatio, resourceRatio float64) (ExtendedEstimatorBuilder, error) {
	switch name {
	case BinpackingEstimatorName:
		return func(
			predicateChecker simulator.PredicateChecker,
			clusterSnapshot simulator.ClusterSnapshot) estimator.Estimator {
			return estimator.NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot)
		}, nil
	case ResourceEstimatorName:
		return func(predicateChecker simulator.PredicateChecker,
			clusterSnapshot simulator.ClusterSnapshot,
		) estimator.Estimator {
			return NewClusterResourceEstimator(predicateChecker, clusterSnapshot,
				cpuRatio, memRatio, resourceRatio)
		}, nil
	}
	return nil, fmt.Errorf("unknown estimator: %s", name)
}
