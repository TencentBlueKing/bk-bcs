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
	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// actionBatch groups consecutive actions that share the same execution mode.
type actionBatch struct {
	actions    []drv1alpha1.Action
	perCluster bool
}

// groupActionBatches splits a workflow's action list into batches by ClusterExecutionMode.
// Consecutive PerCluster actions are merged into a single batch for concurrent per-cluster execution.
// Global actions each form their own batch (serial execution preserved).
func groupActionBatches(actions []drv1alpha1.Action) []actionBatch {
	if len(actions) == 0 {
		return nil
	}

	var batches []actionBatch
	var current *actionBatch

	for i := range actions {
		pc := isPerClusterMode(&actions[i])

		if current == nil || current.perCluster != pc {
			if current != nil {
				batches = append(batches, *current)
			}
			current = &actionBatch{
				perCluster: pc,
				actions:    []drv1alpha1.Action{actions[i]},
			}
		} else if pc {
			current.actions = append(current.actions, actions[i])
		} else {
			batches = append(batches, *current)
			current = &actionBatch{
				perCluster: false,
				actions:    []drv1alpha1.Action{actions[i]},
			}
		}
	}
	if current != nil {
		batches = append(batches, *current)
	}

	return batches
}

// aggregateClusterStatuses computes the overall phase from per-cluster statuses.
// Priority: any Failed → Failed; any Running → Running;
// all terminal (Succeeded/Skipped mix): all Skipped → Skipped, else Succeeded;
// otherwise Pending.
func aggregateClusterStatuses(statuses []drv1alpha1.ClusterActionStatus) string {
	if len(statuses) == 0 {
		return drv1alpha1.PhasePending
	}

	hasFailed := false
	hasRunning := false
	hasPending := false
	allSkipped := true

	for i := range statuses {
		switch statuses[i].Phase {
		case drv1alpha1.PhaseFailed:
			hasFailed = true
			allSkipped = false
		case drv1alpha1.PhaseRunning:
			hasRunning = true
			allSkipped = false
		case drv1alpha1.PhaseSucceeded:
			allSkipped = false
		case drv1alpha1.PhaseSkipped:
			// allSkipped stays true
		default:
			hasPending = true
			allSkipped = false
		}
	}

	if hasFailed {
		return drv1alpha1.PhaseFailed
	}
	if hasRunning {
		return drv1alpha1.PhaseRunning
	}
	if hasPending {
		return drv1alpha1.PhasePending
	}
	if allSkipped {
		return drv1alpha1.PhaseSkipped
	}
	return drv1alpha1.PhaseSucceeded
}
