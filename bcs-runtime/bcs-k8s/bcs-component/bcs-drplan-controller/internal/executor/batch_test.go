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
	"testing"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

func TestGroupActionBatches(t *testing.T) {
	t.Run("all Global actions form one batch per action", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a1", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true},
			{Name: "a2", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true},
		}
		batches := groupActionBatches(actions)
		if len(batches) != 2 {
			t.Fatalf("batches count = %d, want 2", len(batches))
		}
		for _, b := range batches {
			if b.perCluster {
				t.Error("expected Global batch, got PerCluster")
			}
		}
	})

	t.Run("consecutive PerCluster actions grouped into one batch", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "hook1", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true, ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster},
			{Name: "hook2", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true, ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster},
		}
		batches := groupActionBatches(actions)
		if len(batches) != 1 {
			t.Fatalf("batches count = %d, want 1", len(batches))
		}
		if !batches[0].perCluster {
			t.Error("expected PerCluster batch")
		}
		if len(batches[0].actions) != 2 {
			t.Errorf("batch actions count = %d, want 2", len(batches[0].actions))
		}
	})

	t.Run("mixed Global and PerCluster creates alternating batches", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "hook1", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true, ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster},
			{Name: "hook2", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true, ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster},
			{Name: "main", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true},
			{Name: "post-hook", Type: drv1alpha1.ActionTypeSubscription, WaitReady: true, ClusterExecutionMode: drv1alpha1.ClusterExecutionModePerCluster},
		}
		batches := groupActionBatches(actions)
		if len(batches) != 3 {
			t.Fatalf("batches count = %d, want 3", len(batches))
		}
		if !batches[0].perCluster || len(batches[0].actions) != 2 {
			t.Errorf("batch[0]: perCluster=%v actions=%d, want true/2", batches[0].perCluster, len(batches[0].actions))
		}
		if batches[1].perCluster || len(batches[1].actions) != 1 {
			t.Errorf("batch[1]: perCluster=%v actions=%d, want false/1", batches[1].perCluster, len(batches[1].actions))
		}
		if !batches[2].perCluster || len(batches[2].actions) != 1 {
			t.Errorf("batch[2]: perCluster=%v actions=%d, want true/1", batches[2].perCluster, len(batches[2].actions))
		}
	})

	t.Run("empty actions returns empty batches", func(t *testing.T) {
		batches := groupActionBatches(nil)
		if len(batches) != 0 {
			t.Errorf("batches count = %d, want 0", len(batches))
		}
	})

	t.Run("backward compat: all legacy actions without clusterExecutionMode", func(t *testing.T) {
		actions := []drv1alpha1.Action{
			{Name: "a1", Type: drv1alpha1.ActionTypeSubscription},
			{Name: "a2", Type: drv1alpha1.ActionTypeJob},
			{Name: "a3", Type: drv1alpha1.ActionTypeHTTP},
		}
		batches := groupActionBatches(actions)
		if len(batches) != 3 {
			t.Fatalf("batches count = %d, want 3", len(batches))
		}
		for i, b := range batches {
			if b.perCluster {
				t.Errorf("batch[%d]: expected Global, got PerCluster", i)
			}
		}
	})
}

func TestAggregateClusterStatuses(t *testing.T) {
	t.Run("all succeeded returns Succeeded", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhaseSucceeded},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhaseSucceeded},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhaseSucceeded {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhaseSucceeded)
		}
	})

	t.Run("any failed returns Failed", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhaseSucceeded},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhaseFailed},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhaseFailed {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhaseFailed)
		}
	})

	t.Run("any running with no failed returns Running", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhaseSucceeded},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhaseRunning},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhaseRunning {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhaseRunning)
		}
	})

	t.Run("all pending returns Pending", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhasePending},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhasePending},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhasePending {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhasePending)
		}
	})

	t.Run("empty statuses returns Pending", func(t *testing.T) {
		phase := aggregateClusterStatuses(nil)
		if phase != drv1alpha1.PhasePending {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhasePending)
		}
	})

	t.Run("all skipped returns Skipped", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhaseSkipped},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhaseSkipped},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhaseSkipped {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhaseSkipped)
		}
	})

	t.Run("mixed succeeded and skipped returns Succeeded", func(t *testing.T) {
		statuses := []drv1alpha1.ClusterActionStatus{
			{Cluster: "ns/c1", Phase: drv1alpha1.PhaseSucceeded},
			{Cluster: "ns/c2", Phase: drv1alpha1.PhaseSkipped},
		}
		phase := aggregateClusterStatuses(statuses)
		if phase != drv1alpha1.PhaseSucceeded {
			t.Errorf("phase = %q, want %q", phase, drv1alpha1.PhaseSucceeded)
		}
	})
}
