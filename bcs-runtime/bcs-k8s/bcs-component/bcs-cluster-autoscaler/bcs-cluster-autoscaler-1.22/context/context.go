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

// Package context xxx
package context

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	processor_callbacks "k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"

	estimatorinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/estimator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

// Context contains user-configurable constant and configuration-related objects passed to
// scale up/scale down functions.
type Context struct {
	*context.AutoscalingContext
	ExtendedEstimatorBuilder estimatorinternal.ExtendedEstimatorBuilder
}

// NewAutoscalingContext returns an autoscaling context from all the necessary parameters passed via arguments
func NewAutoscalingContext(options scalingconfig.Options,
	predicateChecker simulator.PredicateChecker,
	clusterSnapshot simulator.ClusterSnapshot,
	autoscalingKubeClients *context.AutoscalingKubeClients,
	cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy,
	estimatorBuilder estimatorinternal.ExtendedEstimatorBuilder,
	processorCallbacks processor_callbacks.ProcessorCallbacks) *Context {
	return &Context{
		AutoscalingContext: &context.AutoscalingContext{
			AutoscalingOptions:     options.AutoscalingOptions,
			CloudProvider:          cloudProvider,
			AutoscalingKubeClients: *autoscalingKubeClients,
			PredicateChecker:       predicateChecker,
			ClusterSnapshot:        clusterSnapshot,
			ExpanderStrategy:       expanderStrategy,
			ProcessorCallbacks:     processorCallbacks,
		},
		ExtendedEstimatorBuilder: estimatorBuilder,
	}
}
