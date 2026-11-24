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

	// ========== NodeGroupSet 相关字段 ==========

	// TriggeredNodeGroupSets 当前触发扩容的 NodeGroupSet 列表(按优先级排序)
	TriggeredNodeGroupSets []string

	// SetResourceSnapshot 当前的 Set 资源快照,用于决策和性能优化
	// 类型定义在 core 包中,这里使用 interface{} 避免循环依赖
	SetResourceSnapshot interface{}

	// NodeGroupSetManager NodeGroupSet 管理器引用
	// 类型定义在 core 包中,这里使用 interface{} 避免循环依赖
	NodeGroupSetManager interface{}

	// NodeToGroupCache 节点到 NodeGroup 的映射缓存
	// 用于避免重复调用 CloudProvider.NodeGroupForNode() API
	// Key: Node Name, Value: NodeGroup ID
	// 在每个 CA 循环开始时构建,被多个组件共享使用
	NodeToGroupCache map[string]string

	// ========== 全局 Buffer Ratio (用于不属于任何 Set 的 NodeGroup) ==========
	// GlobalBufferedCPURatio 全局 CPU buffer ratio (来自命令行参数 --cpu-ratio)
	GlobalBufferedCPURatio float64
	// GlobalBufferedMemRatio 全局 Memory buffer ratio (来自命令行参数 --mem-ratio)
	GlobalBufferedMemRatio float64
	// GlobalBufferedResourceRatio 全局 Resource buffer ratio (来自命令行参数 --resource-ratio)
	GlobalBufferedResourceRatio float64
}

// NewAutoscalingContext returns an autoscaling context from all the necessary parameters passed via arguments
func NewAutoscalingContext(options scalingconfig.Options, predicateChecker *simulator.PredicateChecker,
	autoscalingKubeClients *context.AutoscalingKubeClients, cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy, estimatorBuilder estimatorinternal.ExtendedEstimatorBuilder,
	processorCallbacks processor_callbacks.ProcessorCallbacks) *Context {
	return &Context{
		AutoscalingContext: &context.AutoscalingContext{
			AutoscalingOptions:     options.AutoscalingOptions,
			CloudProvider:          cloudProvider,
			AutoscalingKubeClients: *autoscalingKubeClients,
			PredicateChecker:       predicateChecker,
			ExpanderStrategy:       expanderStrategy,
			ProcessorCallbacks:     processorCallbacks,
		},
		ExtendedEstimatorBuilder: estimatorBuilder,
	}
}
