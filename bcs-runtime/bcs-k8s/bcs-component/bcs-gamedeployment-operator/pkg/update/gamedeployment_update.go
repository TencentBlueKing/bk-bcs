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

package update

import (
	"context"
	"fmt"
	"sort"
	"time"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdcore "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/core"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util/canary"
	hooklister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/postinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/preinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/hotpatchupdate"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/requeueduration"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// Interface for managing pods updating.
type Interface interface {
	Manage(deploy, updateDeploy *gdv1alpha1.GameDeployment,
		updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
		pods []*v1.Pod,
		newStatus *gdv1alpha1.GameDeploymentStatus,
	) (time.Duration, error)
}

// New create pod update interface for gamedeployment
func New(kubeClient clientset.Interface, recorder record.EventRecorder, scaleExp expectations.ScaleExpectations,
	updateExp expectations.UpdateExpectations, hookRunLister hooklister.HookRunLister,
	hookTemplateLister hooklister.HookTemplateLister, preDeleteControl predelete.PreDeleteInterface,
	preInplaceControl preinplace.PreInplaceInterface, postInplaceControl postinplace.PostInplaceInterface,
	metrics *gdmetrics.Metrics) Interface {
	return &realControl{
		inPlaceControl:     inplaceupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
		hotPatchControl:    hotpatchupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
		kubeClient:         kubeClient,
		recorder:           recorder,
		scaleExp:           scaleExp,
		updateExp:          updateExp,
		hookRunLister:      hookRunLister,
		hookTemplateLister: hookTemplateLister,
		preDeleteControl:   preDeleteControl,
		preInplaceControl:  preInplaceControl,
		postInplaceControl: postInplaceControl,
		metrics:            metrics,
	}
}

type realControl struct {
	kubeClient         clientset.Interface
	inPlaceControl     inplaceupdate.Interface
	hotPatchControl    hotpatchupdate.Interface
	recorder           record.EventRecorder
	scaleExp           expectations.ScaleExpectations
	updateExp          expectations.UpdateExpectations
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
	preDeleteControl   predelete.PreDeleteInterface
	preInplaceControl  preinplace.PreInplaceInterface
	postInplaceControl postinplace.PostInplaceInterface
	metrics            *gdmetrics.Metrics
}

func (c *realControl) Manage(deploy, updateDeploy *gdv1alpha1.GameDeployment,
	updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
	pods []*v1.Pod, newStatus *gdv1alpha1.GameDeploymentStatus,
) (time.Duration, error) {

	requeueDuration := requeueduration.Duration{}
	coreControl := gdcore.New(updateDeploy)

	if updateDeploy.Spec.UpdateStrategy.Paused {
		return requeueDuration.Get(), nil
	}

	// 1. find currently updated and not-ready count and all pods waiting to update
	var waitUpdateIndexes []int
	for i := range pods {
		if coreControl.IsPodUpdatePaused(pods[i]) {
			continue
		}

		if res := c.inPlaceControl.Refresh(pods[i], coreControl.GetUpdateOptions()); res.RefreshErr != nil {
			klog.Errorf("GameDeployment %s/%s failed to update pod %s condition for inplace: %v",
				updateDeploy.Namespace, updateDeploy.Name, pods[i].Name, res.RefreshErr)
			return requeueDuration.Get(), res.RefreshErr
		} else if res.DelayDuration > 0 {
			requeueDuration.Update(res.DelayDuration)
		}

		if util.GetPodRevision(pods[i]) != updateRevision.Name {
			waitUpdateIndexes = append(waitUpdateIndexes, i)
		}
	}

	// resync post inplace hook status
	for _, pod := range pods {
		err := c.postInplaceControl.UpdatePostInplaceHook(updateDeploy, pod, newStatus, gdv1alpha1.GameDeploymentInstanceID)
		if err != nil {
			c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedResyncPostHookRun",
				"failed to resync post hook for pod %s, error: %v", pod.Name, err)
		}
	}

	// 2. sort all pods waiting to update
	waitUpdateIndexes = sortUpdateIndexes(coreControl, pods, waitUpdateIndexes)

	// 3. calculate max count of pods can update
	needToUpdateCount := calculateUpdateCount(updateDeploy, coreControl, updateDeploy.Spec.UpdateStrategy, updateDeploy.Spec.MinReadySeconds,
		int(*updateDeploy.Spec.Replicas), waitUpdateIndexes, pods)
	if needToUpdateCount < len(waitUpdateIndexes) {
		waitUpdateIndexes = waitUpdateIndexes[:needToUpdateCount]
	}

	// 4. update pods
	for _, idx := range waitUpdateIndexes {
		pod := pods[idx]
		if duration, err := c.updatePod(updateDeploy, coreControl, updateRevision, revisions, pod, newStatus); err != nil {
			return requeueDuration.Get(), err
		} else if duration > 0 {
			requeueDuration.Update(duration)
		}
	}

	return requeueDuration.Get(), nil
}

func sortUpdateIndexes(coreControl gdcore.Control, pods []*v1.Pod, waitUpdateIndexes []int) []int {
	// Sort Pods with default sequence
	sort.SliceStable(waitUpdateIndexes, coreControl.GetPodsSortFunc(pods, waitUpdateIndexes))
	return waitUpdateIndexes
}

func calculateUpdateCount(deploy *gdv1alpha1.GameDeployment, coreControl gdcore.Control, strategy gdv1alpha1.GameDeploymentUpdateStrategy,
	minReadySeconds int32, totalReplicas int, waitUpdateIndexes []int, pods []*v1.Pod) int {

	currentPartition := canary.GetCurrentPartition(deploy)
	if len(waitUpdateIndexes)-int(currentPartition) <= 0 {
		return 0
	}
	waitUpdateIndexes = waitUpdateIndexes[:(len(waitUpdateIndexes) - int(currentPartition))]

	roundUp := true
	if strategy.MaxSurge != nil {
		maxSurge, _ := intstrutil.GetValueFromIntOrPercent(strategy.MaxSurge, totalReplicas, true)
		roundUp = maxSurge == 0
	}
	maxUnavailable, _ := intstrutil.GetValueFromIntOrPercent(
		intstrutil.ValueOrDefault(strategy.MaxUnavailable, intstrutil.FromString(gdv1alpha1.DefaultGameDeploymentMaxUnavailable)), totalReplicas, roundUp)
	usedSurge := len(pods) - totalReplicas

	var notReadyCount, updateCount int
	for _, p := range pods {
		if !coreControl.IsPodUpdateReady(p, minReadySeconds) {
			notReadyCount++
		}
	}
	for _, i := range waitUpdateIndexes {
		if coreControl.IsPodUpdateReady(pods[i], minReadySeconds) {
			if notReadyCount >= (maxUnavailable + usedSurge) {
				break
			} else {
				notReadyCount++
			}
		}
		updateCount++
	}

	return updateCount
}

func (c *realControl) updatePod(deploy *gdv1alpha1.GameDeployment, coreControl gdcore.Control,
	updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
	pod *v1.Pod, newStatus *gdv1alpha1.GameDeploymentStatus,
) (time.Duration, error) {
	var oldRevision *apps.ControllerRevision
	for _, r := range revisions {
		if r.Name == util.GetPodRevision(pod) {
			oldRevision = r
			break
		}
	}

	switch deploy.Spec.UpdateStrategy.Type {
	case gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType:
		if deploy.Spec.PreInplaceUpdateStrategy.Hook != nil {
			klog.V(2).Infof("PreInplace Hook check for inplace update the pod %s/%s now.", pod.Name, pod.Namespace)

			canInplace, err := c.preInplaceControl.CheckInplace(
				deploy, pod, &deploy.Spec.Template, newStatus, gdv1alpha1.GameDeploymentInstanceID)
			if err != nil {
				return 0, err
			}
			if canInplace {
				if pod.Status.Phase != v1.PodRunning {
					klog.V(2).Infof("Pod %s/%s is not running, skip PreInplace Hook run checking.", pod.Name, pod.Namespace)
				} else if deploy.Spec.PreInplaceUpdateStrategy.Hook != nil {
					klog.V(2).Infof("PreInplace Hook run successfully, inplace update the pod %s/%s now.", pod.Name, pod.Namespace)
				}
			} else {
				klog.V(2).Infof("PreInplace Hook not completed, can't inplace update the pod %s/%s now.", pod.Name, pod.Namespace)
				return 0, nil
			}

		} else {
			klog.V(2).Infof("PreDelete Hook check for inplace update the pod %s/%s now.", pod.Name, pod.Namespace)

			canDelete, err := c.preDeleteControl.CheckDelete(deploy, pod, newStatus, gdv1alpha1.GameDeploymentInstanceID)
			if err != nil {
				return 0, err
			}
			if canDelete {
				if deploy.Spec.PreDeleteUpdateStrategy.Hook != nil {
					klog.V(2).Infof("PreDelete Hook run successfully, inplace update the pod %s/%s now.", pod.Name, pod.Namespace)
				}
			} else {
				klog.V(2).Infof("PreDelete Hook not completed, can't inplace update the pod %s/%s now.", pod.Name, pod.Namespace)
				return 0, nil
			}
		}

		startTime := time.Now()
		res := c.inPlaceControl.Update(pod, oldRevision, updateRevision, coreControl.GetUpdateOptions())

		if res.InPlaceUpdate {
			if res.UpdateErr == nil {
				c.recorder.Eventf(deploy, v1.EventTypeNormal, "SuccessfulUpdatePodInPlace",
					"successfully update pod %s in-place", pod.Name)
				c.metrics.CollectPodUpdateDurations(util.GetControllerKey(deploy), "success",
					string(gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType), time.Since(startTime))
				c.updateExp.ExpectUpdated(util.GetControllerKey(deploy), updateRevision.Name, pod)

				// create post inplace hook
				newPod, err := c.kubeClient.CoreV1().Pods(pod.Namespace).Get(context.TODO(), pod.Name, metav1.GetOptions{})
				if err != nil {
					klog.Warningf("Cannot get pod %s/%s", pod.Namespace, pod.Name)
					return res.DelayDuration, nil
				}
				created, err := c.postInplaceControl.CreatePostInplaceHook(deploy, newPod, newStatus,
					gdv1alpha1.GameDeploymentInstanceID)
				if err != nil {
					c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedCreatePostHookRun",
						"failed to create post hook for pod %s, error: %v", pod.Name, err)
				} else if created {
					c.recorder.Eventf(deploy, v1.EventTypeNormal, "SuccessfulCreatePostHookRun",
						"successfully create post hook for pod %s", pod.Name)
				} else {
					c.recorder.Eventf(deploy, v1.EventTypeNormal, "PostHookRunExisted",
						"post hook for pod %s has been existed", pod.Name)
				}
				return res.DelayDuration, nil
			}

			c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedUpdatePodInPlace", "failed to update pod %s in-place: %v", pod.Name, res.UpdateErr)
			c.metrics.CollectPodUpdateDurations(util.GetControllerKey(deploy), "failure",
				string(gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType), time.Since(startTime))
			return res.DelayDuration, res.UpdateErr

		}

		err := fmt.Errorf("find Pod %s update strategy is InPlace, but the diff not only contains replace operation of spec.containers[x].image", pod)
		c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedUpdatePodInPlace", "find Pod %s update strategy is InPlace but can not update in-place: %v", pod.Name, err)
		klog.Warningf("GameDeployment %s/%s can not update Pod %s in-place: %+v", deploy.Namespace, deploy.Name, pod.Name, err)
		return res.DelayDuration, err
	case gdv1alpha1.RollingGameDeploymentUpdateStrategyType:
		canDelete, err := c.preDeleteControl.CheckDelete(deploy, pod, newStatus, gdv1alpha1.GameDeploymentInstanceID)
		if err != nil {
			return 0, err
		}
		if canDelete {
			if deploy.Spec.PreDeleteUpdateStrategy.Hook != nil {
				klog.V(2).Infof("PreDelete Hook run successfully, rolling update the pod %s/%s now.", pod.Name, pod.Namespace)
			}
		} else {
			klog.V(2).Infof("PreDelete Hook not completed, can't rolling update the pod %s/%s now.", pod.Name, pod.Namespace)
			return 0, nil
		}

		klog.V(2).Infof("GameDeployment %s/%s deleting Pod %s for update %s", deploy.Namespace, deploy.Name, pod.Name, updateRevision.Name)
		c.scaleExp.ExpectScale(util.GetControllerKey(deploy), expectations.Delete, pod.Name)
		startTime := time.Now()
		if err := c.kubeClient.CoreV1().Pods(deploy.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{}); err != nil {
			c.scaleExp.ObserveScale(util.GetControllerKey(deploy), expectations.Delete, pod.Name)
			c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedUpdatePodReCreate",
				"failed to delete pod %s for update: %v", pod.Name, err)
			c.metrics.CollectPodDeleteDurations(util.GetControllerKey(deploy), "failure", time.Since(startTime))
			return 0, err
		}

		c.recorder.Eventf(deploy, v1.EventTypeNormal, "SuccessfulUpdatePodReCreate",
			"successfully delete pod %s for update", pod.Name)
		c.metrics.CollectPodDeleteDurations(util.GetControllerKey(deploy), "success", time.Since(startTime))
		return 0, nil

	case gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType:
		startTime := time.Now()
		err := c.hotPatchControl.Update(pod, oldRevision, updateRevision)
		if err != nil {
			c.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedUpdatePodHotPatch", "failed to update pod %s hot-patch: %v", pod.Name, err)
			c.metrics.CollectPodUpdateDurations(util.GetControllerKey(deploy), "failure",
				string(gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType), time.Since(startTime))
			return 0, err
		}
		c.recorder.Eventf(deploy, v1.EventTypeNormal, "SuccessfulUpdatePodHotPatch", "successfully update pod %s hot-patch", pod.Name)
		c.metrics.CollectPodUpdateDurations(util.GetControllerKey(deploy), "success",
			string(gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType), time.Since(startTime))
		c.updateExp.ExpectUpdated(util.GetControllerKey(deploy), updateRevision.Name, pod)
		return 0, nil
	}

	return 0, fmt.Errorf("invalid update strategy type")
}
