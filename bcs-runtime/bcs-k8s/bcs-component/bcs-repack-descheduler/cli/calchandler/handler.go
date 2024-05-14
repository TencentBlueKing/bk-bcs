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

// Package calchandler xx
package calchandler

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/cachemanager"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/pkg/controller/calculator/remote"
)

// CalculatorHandler defines the handler of calculator
type CalculatorHandler struct {
	ctx          context.Context
	cacheManager cachemanager.CacheInterface
}

// NewCalculatorHandler create the calculator handler instance
func NewCalculatorHandler(ctx context.Context, m cachemanager.CacheInterface) *CalculatorHandler {
	return &CalculatorHandler{
		ctx:          ctx,
		cacheManager: m,
	}
}

func (h *CalculatorHandler) getResultPlan(request *calculator.CalculateConvergeRequest) (*calculator.ResultPlan, error) {
	migrator := remote.NewCalculatorRemote(options.GlobalConfigHandler().GetOptions())
	blog.Infof("Calculator requesting...")
	resultPlan, err := migrator.Calculate(h.ctx, request)
	if err != nil {
		return nil, errors.Wrapf(err, "calculator request failed")
	}
	return &resultPlan, nil
}

func (h *CalculatorHandler) buildPodMap(request *calculator.CalculateConvergeRequest,
	resultPlan *calculator.ResultPlan) (original map[string]*calculator.PodItem,
	optimized map[string]*calculator.PodItem) {
	originalPod := request.Original.Pods
	originalPodMap := make(map[string]*calculator.PodItem)
	for _, pod := range originalPod {
		originalPodMap[pod.Item] = pod
	}
	optimizedPodMap := make(map[string]*calculator.PodItem)
	for k, v := range originalPodMap {
		podItem := *v
		newPodItem := podItem
		optimizedPodMap[k] = &newPodItem
	}
	for _, plan := range resultPlan.Plans[0].MigratePlan {
		originalPodItem, ok := originalPodMap[plan.Item]
		if !ok {
			blog.Warnf("plan pod '%s' not found in original results", plan.Item)
			continue
		}
		if originalPodItem.Container != plan.From {
			blog.Warnf("plan pod '%s' from '%s' not same", plan.Item, plan.From)
			continue
		}
		optimizedPodMap[plan.Item].Container = plan.To
	}
	return originalPodMap, optimizedPodMap
}

func (h *CalculatorHandler) buildOptimizedNodes(resultPlan *calculator.ResultPlan, optimizedRate *PackingRate) []string {
	migrateFrom := make(map[string]string)
	for _, plan := range resultPlan.Plans[0].MigratePlan {
		migrateFrom[plan.From] = plan.From
	}
	decreaseNodes := make([]string, 0, len(migrateFrom))
	for nodeFrom := range migrateFrom {
		v, ok := optimizedRate.NodePackingRate[nodeFrom]
		if ok && (v.Cpu < 1 || v.Mem < 1) {
			decreaseNodes = append(decreaseNodes, nodeFrom)
		}
	}
	for _, node := range decreaseNodes {
		nodeRate, ok := optimizedRate.NodePackingRate[node]
		if ok {
			optimizedRate.TotalRate.CpuCapacity -= nodeRate.CpuCapacity
			optimizedRate.TotalRate.MemCapacity -= nodeRate.MemCapacity
			optimizedRate.TotalRate.CpuVal -= nodeRate.CpuVal
			optimizedRate.TotalRate.MemVal -= nodeRate.MemVal
		}

		delete(optimizedRate.NodePods, node)
		delete(optimizedRate.NodePackingRate, node)
	}
	optimizedRate.TotalRate.Mem = optimizedRate.TotalRate.MemVal / optimizedRate.TotalRate.MemCapacity * 100
	optimizedRate.TotalRate.Cpu = optimizedRate.TotalRate.CpuVal / optimizedRate.TotalRate.CpuCapacity * 100
	return decreaseNodes
}

// Calc the repack result
func (h *CalculatorHandler) Calc() (*ClusterRate, error) {
	request, err := h.cacheManager.BuildCalculatorRequest(h.ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "build calculator failed")
	}
	resultPlan, err := h.getResultPlan(request)
	if err != nil {
		return nil, errors.Wrapf(err, "get result plan failed")
	}

	originalPM, optimizedPM := h.buildPodMap(request, resultPlan)
	optimizedRate := h.calcPackingRate(request.Original.Nodes, optimizedPM)
	optimizedNodes := h.buildOptimizedNodes(resultPlan, optimizedRate)
	originalRate := h.calcPackingRate(request.Original.Nodes, originalPM)
	return &ClusterRate{
		OptimizedNodes: optimizedNodes,
		OriginalRate:   originalRate,
		OptimizedRate:  optimizedRate,
	}, nil
}

func (h *CalculatorHandler) calcPackingRate(nodes []*calculator.NodeItem,
	podMap map[string]*calculator.PodItem) *PackingRate {
	rate := &PackingRate{
		NodePods:        make(map[string][]*calculator.PodItem),
		NodePackingRate: make(map[string]*RateObj),
	}

	var nodeTotalMem, nodeTotalCpu float64
	var podTotalMem, podTotalCpu float64
	nodePods := make(map[string][]*calculator.PodItem)
	for _, podItem := range podMap {
		nodePods[podItem.Container] = append(nodePods[podItem.Container], podItem)
	}
	for _, node := range nodes {
		podItems, ok := nodePods[node.Container]
		if !ok {
			blog.Warnf("node '%s' have no pods", node.Container)
			continue
		}
		rate.NodePods[node.Container] = podItems

		var nodePodMem, nodePodCpu float64
		for _, item := range podItems {
			nodePodMem += item.Index1
			nodePodCpu += item.Index2
		}
		rate.NodePackingRate[node.Container] = &RateObj{
			MemVal:      nodePodMem,
			CpuVal:      nodePodCpu,
			MemCapacity: node.Index1,
			CpuCapacity: node.Index2,
			Mem:         nodePodMem / node.Index1 * 100,
			Cpu:         nodePodCpu / node.Index2 * 100,
		}
		nodeTotalMem += node.Index1
		nodeTotalCpu += node.Index2
		podTotalMem += nodePodMem
		podTotalCpu += nodePodCpu
	}
	rate.TotalRate = &RateObj{
		MemVal:      podTotalMem,
		CpuVal:      podTotalCpu,
		MemCapacity: nodeTotalMem,
		CpuCapacity: nodeTotalCpu,
		Mem:         podTotalMem / nodeTotalMem * 100,
		Cpu:         podTotalCpu / nodeTotalCpu * 100,
	}
	return rate
}
