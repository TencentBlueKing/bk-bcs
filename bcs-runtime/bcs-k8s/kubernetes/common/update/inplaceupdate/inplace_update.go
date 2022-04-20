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

package inplaceupdate

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

// UpdateOptions is the option of update
type UpdateOptions struct {
	GracePeriodSeconds int32
}

// RefreshResult is the result of refresh
type RefreshResult struct {
	RefreshErr    error
	DelayDuration time.Duration
}

// UpdateResult is the result of update
type UpdateResult struct {
	InPlaceUpdate bool
	UpdateErr     error
	DelayDuration time.Duration
}

// Interface for managing pods in-place update.
type Interface interface {
	Refresh(pod *v1.Pod, opts *UpdateOptions) RefreshResult
	Update(pod *v1.Pod, oldRevision, newRevision *apps.ControllerRevision, opts *UpdateOptions) UpdateResult
}

type realControl struct {
	adp         update.Adapter
	revisionKey string

	// just for test
	now func() metav1.Time
}

// NewForTypedClient constructs realControl
func NewForTypedClient(c clientset.Interface, revisionKey string) Interface {
	return &realControl{adp: &update.AdapterTypedClient{Client: c}, revisionKey: revisionKey, now: metav1.Now}
}

// Refresh refreshs condition and grace period of inplace
func (c *realControl) Refresh(pod *v1.Pod, opts *UpdateOptions) RefreshResult {
	if err := c.refreshCondition(pod, opts); err != nil {
		return RefreshResult{RefreshErr: err}
	}

	var delayDuration time.Duration
	var err error
	if pod.Annotations[InPlaceUpdateGraceKey] != "" {
		if delayDuration, err = c.finishGracePeriod(pod, opts); err != nil {
			return RefreshResult{RefreshErr: err}
		}
	}

	return RefreshResult{DelayDuration: delayDuration}
}

// Update executes inplace update
func (c *realControl) Update(pod *v1.Pod, oldRevision, newRevision *apps.ControllerRevision,
	opts *UpdateOptions) UpdateResult {
	// 1. calculate inplace update spec
	spec := update.CalculateInPlaceUpdateSpec(oldRevision, newRevision)

	if spec == nil {
		return UpdateResult{}
	}
	if opts != nil && opts.GracePeriodSeconds > 0 {
		spec.GraceSeconds = opts.GracePeriodSeconds
	}

	// 2. update condition for pod with readiness-gate
	if containsReadinessGate(pod) {
		newCondition := v1.PodCondition{
			Type:               InPlaceUpdateReady,
			LastTransitionTime: c.now(),
			Status:             v1.ConditionFalse,
			Reason:             "StartInPlaceUpdate",
		}
		if err := c.updateCondition(pod, newCondition); err != nil {
			return UpdateResult{InPlaceUpdate: true, UpdateErr: err}
		}
	}

	// 3. update container images
	if err := c.updatePodInPlace(pod, spec, opts); err != nil {
		return UpdateResult{InPlaceUpdate: true, UpdateErr: err}
	}

	var delayDuration time.Duration
	if opts != nil && opts.GracePeriodSeconds > 0 {
		delayDuration = time.Second * time.Duration(opts.GracePeriodSeconds)
	}
	return UpdateResult{InPlaceUpdate: true, DelayDuration: delayDuration}
}

func (c *realControl) refreshCondition(pod *v1.Pod, opts *UpdateOptions) error {
	// no need to update condition because of no readiness-gate
	if !containsReadinessGate(pod) {
		return nil
	}

	// in-place updating has not completed yet
	if checkErr := CheckInPlaceUpdateCompleted(pod); checkErr != nil {
		klog.V(6).Infof("Check Pod %s/%s in-place update not completed yet: %v", pod.Namespace, pod.Name, checkErr)
		return nil
	}

	// already ready
	if existingCondition := GetCondition(pod); existingCondition != nil && existingCondition.Status == v1.ConditionTrue {
		return nil
	}

	newCondition := v1.PodCondition{
		Type:               InPlaceUpdateReady,
		Status:             v1.ConditionTrue,
		LastTransitionTime: c.now(),
	}
	return c.updateCondition(pod, newCondition)
}

func (c *realControl) updateCondition(pod *v1.Pod, condition v1.PodCondition) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		clone, err := c.adp.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			return err
		}

		setPodCondition(clone, condition)
		return c.adp.UpdatePodStatus(clone)
	})
}

func (c *realControl) finishGracePeriod(pod *v1.Pod, opts *UpdateOptions) (time.Duration, error) {
	var delayDuration time.Duration
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		clone, err := c.adp.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			return err
		}

		spec := update.UpdateSpec{}
		updateSpecJSON, ok := clone.Annotations[InPlaceUpdateGraceKey]
		if !ok {
			return nil
		}
		if jsonErr := json.Unmarshal([]byte(updateSpecJSON), &spec); jsonErr != nil {
			return err
		}
		graceDuration := time.Second * time.Duration(spec.GraceSeconds)

		updateState := InPlaceUpdateState{}
		updateStateJSON, ok := clone.Annotations[InPlaceUpdateStateKey]
		if !ok {
			return fmt.Errorf("pod has %s but %s not found", InPlaceUpdateGraceKey, InPlaceUpdateStateKey)
		}
		if jsonErr := json.Unmarshal([]byte(updateStateJSON), &updateState); jsonErr != nil {
			return nil
		}

		if clone.Labels[c.revisionKey] != spec.Revision {
			// If revision-hash has changed, just drop this GracePeriodSpec and go through the normal update process again.
			delete(clone.Annotations, InPlaceUpdateGraceKey)
		} else {
			if span := time.Since(updateState.UpdateTimestamp.Time); span < graceDuration {
				delayDuration = roundupSeconds(graceDuration - span)
				return nil
			}

			if clone, err = update.PatchUpdateSpecToPod(clone, &spec); err != nil {
				return err
			}
			delete(clone.Annotations, InPlaceUpdateGraceKey)
		}

		return c.adp.UpdatePod(clone)
	})

	return delayDuration, err
}

func (c *realControl) updatePodInPlace(pod *v1.Pod, spec *update.UpdateSpec, opts *UpdateOptions) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		clone, err := c.adp.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			return err
		}

		// update new revision
		if c.revisionKey != "" {
			clone.Labels[c.revisionKey] = spec.Revision
		}
		if clone.Annotations == nil {
			clone.Annotations = map[string]string{}
		}

		// record old containerStatuses
		inPlaceUpdateState := InPlaceUpdateState{
			Revision:              spec.Revision,
			UpdateTimestamp:       c.now(),
			LastContainerStatuses: make(map[string]InPlaceUpdateContainerStatus, len(spec.ContainerImages)),
		}
		for _, c := range clone.Status.ContainerStatuses {
			if _, ok := spec.ContainerImages[c.Name]; ok {
				inPlaceUpdateState.LastContainerStatuses[c.Name] = InPlaceUpdateContainerStatus{
					ImageID: c.ImageID,
				}
			}
		}
		inPlaceUpdateStateJSON, _ := json.Marshal(inPlaceUpdateState)
		clone.Annotations[InPlaceUpdateStateKey] = string(inPlaceUpdateStateJSON)

		if spec.GraceSeconds <= 0 {
			if clone, err = update.PatchUpdateSpecToPod(clone, spec); err != nil {
				return err
			}
			delete(clone.Annotations, InPlaceUpdateGraceKey)
		} else {
			inPlaceUpdateSpecJSON, _ := json.Marshal(spec)
			clone.Annotations[InPlaceUpdateGraceKey] = string(inPlaceUpdateSpecJSON)
		}

		return c.adp.UpdatePod(clone)
	})
}

// CheckInPlaceUpdateCompleted checks whether imageID in pod status has been changed since in-place update.
// If the imageID in containerStatuses has not been changed, we assume that kubelet has not updated
// containers in Pod.
func CheckInPlaceUpdateCompleted(pod *v1.Pod) error {
	inPlaceUpdateState := InPlaceUpdateState{}
	if stateStr, ok := pod.Annotations[InPlaceUpdateStateKey]; !ok {
		return nil
	} else if err := json.Unmarshal([]byte(stateStr), &inPlaceUpdateState); err != nil {
		return err
	}

	// this should not happen, unless someone modified pod revision label
	if inPlaceUpdateState.Revision != pod.Labels[apps.StatefulSetRevisionLabel] {
		return fmt.Errorf("currently revision %s not equal to in-place update revision %s",
			pod.Labels[apps.StatefulSetRevisionLabel], inPlaceUpdateState.Revision)
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if oldStatus, ok := inPlaceUpdateState.LastContainerStatuses[cs.Name]; ok {
			// TODO: we assume that users should not update workload template with new image which
			// actually has the same imageID as the old image
			if oldStatus.ImageID == cs.ImageID {
				return fmt.Errorf("container %s imageID not changed", cs.Name)
			}
			delete(inPlaceUpdateState.LastContainerStatuses, cs.Name)
		}
	}

	if len(inPlaceUpdateState.LastContainerStatuses) > 0 {
		return fmt.Errorf("not found statuses of containers %v", inPlaceUpdateState.LastContainerStatuses)
	}

	return nil
}

func containsReadinessGate(pod *v1.Pod) bool {
	for _, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == InPlaceUpdateReady {
			return true
		}
	}
	return false
}

// GetCondition returns the InPlaceUpdateReady condition in Pod.
func GetCondition(pod *v1.Pod) *v1.PodCondition {
	return getCondition(pod, InPlaceUpdateReady)
}

func getCondition(pod *v1.Pod, cType v1.PodConditionType) *v1.PodCondition {
	for _, c := range pod.Status.Conditions {
		if c.Type == cType {
			return &c
		}
	}
	return nil
}

func setPodCondition(pod *v1.Pod, condition v1.PodCondition) {
	for i, c := range pod.Status.Conditions {
		if c.Type == condition.Type {
			if c.Status != condition.Status {
				pod.Status.Conditions[i] = condition
			}
			return
		}
	}
	pod.Status.Conditions = append(pod.Status.Conditions, condition)
}

func roundupSeconds(d time.Duration) time.Duration {
	if d%time.Second == 0 {
		return d
	}
	return (d/time.Second + 1) * time.Second
}

// InjectReadinessGate injects InPlaceUpdateReady into pod.spec.readinessGates
func InjectReadinessGate(pod *v1.Pod) {
	for _, r := range pod.Spec.ReadinessGates {
		if r.ConditionType == InPlaceUpdateReady {
			return
		}
	}
	pod.Spec.ReadinessGates = append(pod.Spec.ReadinessGates, v1.PodReadinessGate{ConditionType: InPlaceUpdateReady})
}
