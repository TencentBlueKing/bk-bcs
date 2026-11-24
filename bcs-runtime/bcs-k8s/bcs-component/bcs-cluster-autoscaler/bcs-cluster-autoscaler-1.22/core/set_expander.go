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
	klog "k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/scalingconfig"
)

// SetAwareExpander wraps the global expander and applies Set-level expander configuration
type SetAwareExpander struct {
	globalExpander expander.Strategy
	setManager     *NodeGroupSetManager
}

// NewSetAwareExpander creates a new Set-aware expander
func NewSetAwareExpander(globalExpander expander.Strategy, setManager *NodeGroupSetManager) expander.Strategy {
	return &SetAwareExpander{
		globalExpander: globalExpander,
		setManager:     setManager,
	}
}

// BestOption selects the best expansion option from the given options
// It first groups options by Set, then applies Set-level expander if configured,
// otherwise falls back to global expander
func (e *SetAwareExpander) BestOption(options []expander.Option,
	nodeInfo map[string]*schedulerframework.NodeInfo,
) *expander.Option {
	if len(options) == 0 {
		return nil
	}

	// If SetManager is not configured, use global expander
	if e.setManager == nil {
		return e.globalExpander.BestOption(options, nodeInfo)
	}

	// Group options by Set
	setOptions := make(map[string][]expander.Option)
	for _, opt := range options {
		setName := e.setManager.GetNodeGroupSet(opt.NodeGroup.Id())
		if setName == "" {
			setName = "default"
		}
		setOptions[setName] = append(setOptions[setName], opt)
	}

	// For each Set, apply Set-level expander or global expander
	bestPerSet := make(map[string]*expander.Option)
	for setName, opts := range setOptions {
		if len(opts) == 0 {
			continue
		}

		// Get Set configuration
		setConfig, err := e.setManager.GetSetConfig(setName)
		if err != nil || setConfig.Expander == "" {
			// No Set-level expander configured, use global expander
			best := e.globalExpander.BestOption(opts, nodeInfo)
			if best != nil {
				bestPerSet[setName] = best
			}
			continue
		}

		// Apply Set-level expander
		best := e.applySetExpander(setConfig, opts, nodeInfo)
		if best != nil {
			bestPerSet[setName] = best
		}
	}

	// If only one Set has options, return its best
	if len(bestPerSet) == 1 {
		for _, best := range bestPerSet {
			return best
		}
	}

	// If multiple Sets have options, apply global expander to best-per-set
	if len(bestPerSet) > 1 {
		bestOptions := make([]expander.Option, 0, len(bestPerSet))
		for _, best := range bestPerSet {
			bestOptions = append(bestOptions, *best)
		}
		return e.globalExpander.BestOption(bestOptions, nodeInfo)
	}

	return nil
}

// applySetExpander applies the Set-level expander strategy
func (e *SetAwareExpander) applySetExpander(setConfig *scalingconfig.NodeGroupSet,
	options []expander.Option, nodeInfo map[string]*schedulerframework.NodeInfo,
) *expander.Option {
	switch setConfig.Expander {
	case "priority":
		return e.applyPriorityExpander(setConfig, options)
	case "random", "most-pods", "least-waste", "price":
		// For other expanders, use global expander
		// TODO: In the future, create Set-specific instances of these expanders
		klog.V(4).Infof("Set %s uses expander %s, falling back to global expander",
			setConfig.SetName, setConfig.Expander)
		return e.globalExpander.BestOption(options, nodeInfo)
	default:
		klog.Warningf("Set %s has unknown expander %s, using global expander",
			setConfig.SetName, setConfig.Expander)
		return e.globalExpander.BestOption(options, nodeInfo)
	}
}

// applyPriorityExpander applies priority-based selection according to Set configuration
func (e *SetAwareExpander) applyPriorityExpander(setConfig *scalingconfig.NodeGroupSet,
	options []expander.Option,
) *expander.Option {
	if len(setConfig.ExpanderPriorities) == 0 {
		klog.Warningf("Set %s: priority expander configured but no priorities specified", setConfig.SetName)
		return &options[0] // Return first option as fallback
	}

	// Build priority map: NodeGroupID -> priority (lower index = higher priority)
	priorityMap := make(map[string]int)
	for i, ngID := range setConfig.ExpanderPriorities {
		priorityMap[ngID] = i
	}

	// Find the option with highest priority
	var bestOption *expander.Option
	bestPriority := -1

	for i := range options {
		opt := &options[i]
		ngID := opt.NodeGroup.Id()

		// Get priority for this NodeGroup
		pri, hasPriority := priorityMap[ngID]
		if !hasPriority {
			// NodeGroup not in priority list, assign lowest priority
			pri = len(setConfig.ExpanderPriorities)
		}

		// Lower priority value = higher priority
		if bestOption == nil || pri < bestPriority {
			bestOption = opt
			bestPriority = pri
		}
	}

	if bestOption != nil {
		klog.V(4).Infof("Set %s: selected NodeGroup %s (priority %d) from %d options",
			setConfig.SetName, bestOption.NodeGroup.Id(), bestPriority, len(options))
	}

	return bestOption
}
