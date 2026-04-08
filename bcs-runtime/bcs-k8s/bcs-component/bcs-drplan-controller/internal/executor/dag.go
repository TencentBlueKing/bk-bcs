/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package executor

import (
	"fmt"
	"strings"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// hasDependsOn returns true when at least one action specifies a non-empty DependsOn list.
func hasDependsOn(actions []drv1alpha1.Action) bool {
	for i := range actions {
		if len(actions[i].DependsOn) > 0 {
			return true
		}
	}
	return false
}

// buildActionGraph builds a forward adjacency list (dependency → dependents) and
// an in-degree map from the action list. It validates that every DependsOn
// reference points to an existing action name.
// Returns (forward edges, in-degree map, error).
func buildActionGraph(actions []drv1alpha1.Action) (map[string][]string, map[string]int, error) {
	nameSet := make(map[string]struct{}, len(actions))
	for i := range actions {
		if _, dup := nameSet[actions[i].Name]; dup {
			return nil, nil, fmt.Errorf("duplicate action name %q", actions[i].Name)
		}
		nameSet[actions[i].Name] = struct{}{}
	}

	forward := make(map[string][]string, len(actions))
	inDegree := make(map[string]int, len(actions))
	for i := range actions {
		inDegree[actions[i].Name] = 0
	}

	for i := range actions {
		seen := make(map[string]struct{}, len(actions[i].DependsOn))
		for _, dep := range actions[i].DependsOn {
			dep = strings.TrimSpace(dep)
			if _, ok := nameSet[dep]; !ok {
				return nil, nil, fmt.Errorf("unknown action %q in dependsOn of action %q", dep, actions[i].Name)
			}
			if dep == actions[i].Name {
				return nil, nil, fmt.Errorf("action %q depends on itself", actions[i].Name)
			}
			if _, dup := seen[dep]; dup {
				continue
			}
			seen[dep] = struct{}{}
			forward[dep] = append(forward[dep], actions[i].Name)
			inDegree[actions[i].Name]++
		}
	}
	return forward, inDegree, nil
}

// topoSortLayers performs a Kahn-style topological sort and returns actions grouped by layer.
// Actions in the same layer have no inter-dependencies and can execute concurrently.
// Within each layer, actions are ordered by their position in the original actions slice
// to guarantee deterministic status output.
func topoSortLayers(actions []drv1alpha1.Action, forward map[string][]string, inDegree map[string]int) ([][]drv1alpha1.Action, error) {
	actionByName := make(map[string]drv1alpha1.Action, len(actions))
	positionOf := make(map[string]int, len(actions))
	for i := range actions {
		actionByName[actions[i].Name] = actions[i]
		positionOf[actions[i].Name] = i
	}

	// Seed queue with zero-indegree nodes, ordered by original position.
	var queue []string
	for i := range actions {
		if inDegree[actions[i].Name] == 0 {
			queue = append(queue, actions[i].Name)
		}
	}

	var layers [][]drv1alpha1.Action
	visited := 0

	for len(queue) > 0 {
		layer := make([]drv1alpha1.Action, 0, len(queue))
		for _, name := range queue {
			layer = append(layer, actionByName[name])
		}
		sortByPosition(layer, positionOf)
		layers = append(layers, layer)
		visited += len(queue)

		var next []string
		for _, name := range queue {
			for _, dependent := range forward[name] {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					next = append(next, dependent)
				}
			}
		}
		// Sort next layer by original position for deterministic ordering.
		sortNamesByPosition(next, positionOf)
		queue = next
	}

	if visited != len(actions) {
		return nil, fmt.Errorf("cycle detected in action dependsOn graph (%d actions in cycle)", len(actions)-visited)
	}

	return layers, nil
}

func sortByPosition(actions []drv1alpha1.Action, positionOf map[string]int) {
	n := len(actions)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && positionOf[actions[j].Name] < positionOf[actions[j-1].Name]; j-- {
			actions[j], actions[j-1] = actions[j-1], actions[j]
		}
	}
}

func sortNamesByPosition(names []string, positionOf map[string]int) {
	n := len(names)
	for i := 1; i < n; i++ {
		for j := i; j > 0 && positionOf[names[j]] < positionOf[names[j-1]]; j-- {
			names[j], names[j-1] = names[j-1], names[j]
		}
	}
}
