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

package core

import (
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/klog"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	contextinternal "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/context"
)

// SetExpansionPlan represents a scale-up plan for a single NodeGroupSet
type SetExpansionPlan struct {
	SetName       string                      // Name of the Set
	BestOption    *expander.Option            // Best expansion option for this Set
	NodeInfo      *schedulernodeinfo.NodeInfo // Node template
	EstimatedCost float64                     // Estimated cost (if available)
}

// selectBestOptionsPerSet groups expansion options by Set and selects the best option for each Set
func selectBestOptionsPerSet(
	context *contextinternal.Context,
	expansionOptions map[string]expander.Option,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
) map[string]*SetExpansionPlan {
	// Extract NodeGroupSetManager from context
	nodeGroupSetManager := getNodeGroupSetManagerFromContext(context)
	if nodeGroupSetManager == nil {
		// No NodeGroupSetManager, treat all options as one "default" Set
		return createSingleSetPlan(expansionOptions, nodeInfos)
	}

	// Group options by Set
	setOptionsMap := make(map[string][]expander.Option)
	for ngID, option := range expansionOptions {
		setName := nodeGroupSetManager.GetNodeGroupSet(ngID)
		if setName == "" {
			setName = "default"
		}
		setOptionsMap[setName] = append(setOptionsMap[setName], option)
	}

	klog.V(2).Infof("[MultiSet] Expansion options grouped into %d Set(s)", len(setOptionsMap))

	// Select best option for each Set using the expander
	setPlans := make(map[string]*SetExpansionPlan)

	for setName, options := range setOptionsMap {
		if len(options) == 0 {
			continue
		}

		// Apply expander to select best option for this Set
		best := selectBestOptionForSet(context, options, nodeInfos)
		if best == nil {
			klog.V(4).Infof("[MultiSet] No best option selected for Set %s", setName)
			continue
		}

		nodeInfo, found := nodeInfos[best.NodeGroup.Id()]
		if !found {
			klog.Warningf("[MultiSet] No node info for best option %s in Set %s",
				best.NodeGroup.Id(), setName)
			continue
		}

		setPlans[setName] = &SetExpansionPlan{
			SetName:    setName,
			BestOption: best,
			NodeInfo:   nodeInfo,
		}

		klog.V(2).Infof("[MultiSet] Set %s: selected NodeGroup %s with %d nodes",
			setName, best.NodeGroup.Id(), best.NodeCount)
	}

	return setPlans
}

// selectBestOptionForSet selects the best option for a single Set
func selectBestOptionForSet(
	context *contextinternal.Context,
	options []expander.Option,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
) *expander.Option {
	if len(options) == 0 {
		return nil
	}

	if len(options) == 1 {
		return &options[0]
	}

	// Use context's expander strategy
	if context.ExpanderStrategy == nil {
		// Fallback: return first option
		klog.V(4).Infof("[MultiSet] No expander strategy available, using first option")
		return &options[0]
	}

	// Let the expander choose
	return context.ExpanderStrategy.BestOption(options, nodeInfos)
}

// selectSetsToExpand selects which Sets should expand based on resource limits and priorities
func selectSetsToExpand(
	setPlans map[string]*SetExpansionPlan,
	resourceLimitsLeft scaleUpResourcesLimits,
	maxConcurrentSets int,
) []*SetExpansionPlan {
	if len(setPlans) == 0 {
		return nil
	}

	// Collect all plans
	allPlans := make([]*SetExpansionPlan, 0, len(setPlans))
	for _, plan := range setPlans {
		allPlans = append(allPlans, plan)
	}

	klog.V(3).Infof("[MultiSet] Selecting from %d Sets (max concurrent: %d)",
		len(allPlans), maxConcurrentSets)

	// Select Sets respecting resource limits and max concurrent limit
	selectedPlans := make([]*SetExpansionPlan, 0)
	remainingLimits := copyResourceLimits(resourceLimitsLeft)

	for _, plan := range allPlans {
		// Check max concurrent Sets limit
		if maxConcurrentSets > 0 && len(selectedPlans) >= maxConcurrentSets {
			klog.V(3).Infof("[MultiSet] Reached max concurrent Sets limit (%d)", maxConcurrentSets)
			break
		}

		// Check if this Set's expansion fits within remaining resource limits
		if canExpandWithinLimits(plan, remainingLimits) {
			selectedPlans = append(selectedPlans, plan)
			updateRemainingLimits(plan, remainingLimits)
			klog.V(2).Infof("[MultiSet] Selected Set %s", plan.SetName)
		} else {
			klog.V(3).Infof("[MultiSet] Skipped Set %s due to resource limits", plan.SetName)
		}
	}

	if len(selectedPlans) == 0 {
		klog.V(2).Info("[MultiSet] No Sets selected for expansion (resource limits exceeded)")
	} else {
		klog.V(1).Infof("[MultiSet] Selected %d Set(s) for concurrent expansion", len(selectedPlans))
	}

	return selectedPlans
}

// canExpandWithinLimits checks if a Set expansion fits within resource limits
func canExpandWithinLimits(plan *SetExpansionPlan, limits scaleUpResourcesLimits) bool {
	// For now, simplified check - always allow if not explicitly exceeded
	return true
}

// updateRemainingLimits updates remaining resource limits after allocating for a Set
func updateRemainingLimits(plan *SetExpansionPlan, limits scaleUpResourcesLimits) {
	// Placeholder for resource limit updates
}

// copyResourceLimits creates a copy of resource limits
func copyResourceLimits(limits scaleUpResourcesLimits) scaleUpResourcesLimits {
	copy := make(scaleUpResourcesLimits)
	for k, v := range limits {
		copy[k] = v
	}
	return copy
}

// createSingleSetPlan creates a plan for all options as a single "default" Set
func createSingleSetPlan(
	expansionOptions map[string]expander.Option,
	nodeInfos map[string]*schedulernodeinfo.NodeInfo,
) map[string]*SetExpansionPlan {
	if len(expansionOptions) == 0 {
		return nil
	}

	// Pick first option as representative
	var firstOption *expander.Option
	var firstNodeInfo *schedulernodeinfo.NodeInfo

	for ngID, option := range expansionOptions {
		opt := option // Copy to avoid pointer issues
		firstOption = &opt
		firstNodeInfo = nodeInfos[ngID]
		break
	}

	return map[string]*SetExpansionPlan{
		"default": {
			SetName:    "default",
			BestOption: firstOption,
			NodeInfo:   firstNodeInfo,
		},
	}
}

// getNodeGroupSetManagerFromContext extracts NodeGroupSetManager from context
func getNodeGroupSetManagerFromContext(context *contextinternal.Context) *NodeGroupSetManager {
	if context.NodeGroupSetManager == nil {
		return nil
	}

	// Type assertion
	if mgr, ok := context.NodeGroupSetManager.(*NodeGroupSetManager); ok {
		return mgr
	}

	return nil
}
