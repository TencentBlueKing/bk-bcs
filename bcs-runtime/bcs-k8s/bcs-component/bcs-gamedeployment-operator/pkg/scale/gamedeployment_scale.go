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

package scale

import (
	"context"
	"fmt"
	"sync"
	"time"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gdcore "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/core"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	hooklister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	// LengthOfInstanceID is the length of instance-id
	LengthOfInstanceID = 5

	// When batching pod creates, initialBatchSize is the size of the initial batch.
	initialBatchSize = 1

	// PodDeletionCost is the cost of pod's deletion
	PodDeletionCost = "controller.kubernetes.io/pod-deletion-cost"
	// NodeDeletionCost is the cost of node's deletion
	NodeDeletionCost = "io.tencent.bcs.dev/node-deletion-cost"

	// DeletionCostSortMethod is the method when sorting deletion cost. Default is ascend.
	DeletionCostSortMethod = "io.tencent.bcs.dev/pod-deletion-cost-sort-method"
	// CostSortMethodAscend will sort costs in asecnding order
	CostSortMethodAscend = "ascend"
	// CostSortMethodDescend will sort costs in descending order
	CostSortMethodDescend = "descend"
)

// Interface for managing replicas including create and delete pod/pvc.
type Interface interface {
	Manage(
		deploy, currentDeploy, updateDeploy *gdv1alpha1.GameDeployment,
		currentRevision, updateRevision string,
		pods []*v1.Pod,
		allPods []*v1.Pod,
		newStatus *gdv1alpha1.GameDeploymentStatus,
	) (bool, error)
}

// New returns a scale control.
func New(kubeClient clientset.Interface, tkexClient gdclientset.Interface, recorder record.EventRecorder,
	exp expectations.ScaleExpectations, hookRunLister hooklister.HookRunLister,
	hookTemplateLister hooklister.HookTemplateLister, nodeLister corelisters.NodeLister,
	preDeleteControl predelete.PreDeleteInterface, metrics *gdmetrics.Metrics) Interface {
	return &realControl{kubeClient: kubeClient, tkexClient: tkexClient, recorder: recorder, exp: exp,
		hookRunLister: hookRunLister, hookTemplateLister: hookTemplateLister, nodeLister: nodeLister,
		preDeleteControl: preDeleteControl, metrics: metrics}
}

type realControl struct {
	kubeClient         clientset.Interface
	tkexClient         gdclientset.Interface
	recorder           record.EventRecorder
	exp                expectations.ScaleExpectations
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
	nodeLister         corelisters.NodeLister
	preDeleteControl   predelete.PreDeleteInterface
	metrics            *gdmetrics.Metrics
}

func (r *realControl) Manage(
	deploy, currentDeploy, updateDeploy *gdv1alpha1.GameDeployment,
	currentRevision, updateRevision string,
	pods []*v1.Pod,
	allPods []*v1.Pod,
	newStatus *gdv1alpha1.GameDeploymentStatus,
) (bool, error) {

	if updateDeploy.Spec.Replicas == nil {
		klog.Errorf("GameDeployment %s has no spec.Replicas", deploy.Name)
		r.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedScale", "failed to scale: has no spec.Replicas")
		return false, fmt.Errorf("spec.Replicas is nil")
	}

	inject, start, end, err := validateGameDeploymentPodIndex(deploy)
	if err != nil {
		klog.Errorf("GameDeployment %s validateGameDeploymentPodIndex failed: %v", deploy.Name, err)
		r.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedScale", "failed to scale: %v", err)
		return false, err
	}

	controllerKey := util.GetControllerKey(updateDeploy)
	coreControl := gdcore.New(updateDeploy)
	if !coreControl.IsReadyToScale() {
		klog.Warningf("GameDeployment %s skip scaling for not ready to scale", controllerKey)
		return false, nil
	}

	if podsToDelete := getPodsToDelete(updateDeploy, pods); len(podsToDelete) > 0 {
		klog.V(3).Infof("GameDeployment %s begin to delete pods in podsToDelete: %v", controllerKey, podsToDelete)
		return r.deletePods(updateDeploy, podsToDelete, newStatus)
	}

	updatedPods, notUpdatedPods := util.SplitPodsByRevision(pods, updateRevision)

	diff, currentRevDiff := calculateDiffs(updateDeploy, updateRevision == currentRevision, len(pods), len(notUpdatedPods))

	if diff < 0 {
		// total number of this creation
		expectedCreations := diff * -1
		// lack number of current version
		expectedCurrentCreations := 0
		if currentRevDiff < 0 {
			expectedCurrentCreations = currentRevDiff * -1
		}

		klog.V(3).Infof("GameDeployment %s begin to scale out %d pods including %d (current rev)",
			controllerKey, expectedCreations, expectedCurrentCreations)

		// when generate id and index, should take all pods (including terminating pods) into accounts
		// generate available ids
		availableIDs := genAvailableIDs(expectedCreations, allPods)
		availableIndex := genAvailableIndex(inject, start, end, allPods)

		return r.createPods(expectedCreations, expectedCurrentCreations,
			currentDeploy, updateDeploy, currentRevision, updateRevision, availableIDs.List(), availableIndex)

	} else if diff > 0 {
		klog.V(3).Infof("GameDeployment %s begin to scale in %d pods including %d (current rev)",
			controllerKey, diff, currentRevDiff)

		sortMethod := getDeletionCostSortMethod(updateDeploy)
		podsToDelete := choosePodsToDelete(diff, currentRevDiff, notUpdatedPods, updatedPods, sortMethod, r.nodeLister)

		return r.deletePods(updateDeploy, podsToDelete, newStatus)
	}

	return false, nil
}

func (r *realControl) createPods(
	expectedCreations, expectedCurrentCreations int,
	currentGD, updateGD *gdv1alpha1.GameDeployment,
	currentRevision, updateRevision string,
	availableIDs []string, availableIndex []int,
) (bool, error) {
	// new all pods need to create
	coreControl := gdcore.New(updateGD)
	newPods, err := coreControl.NewVersionedPods(currentGD, updateGD, currentRevision, updateRevision,
		expectedCreations, expectedCurrentCreations, availableIDs, availableIndex)
	if err != nil {
		return false, err
	}

	podsCreationChan := make(chan *v1.Pod, len(newPods))
	for _, p := range newPods {
		r.exp.ExpectScale(util.GetControllerKey(updateGD), expectations.Create, p.Name)
		podsCreationChan <- p
	}

	var created bool
	successPodNames := sync.Map{}
	_, err = util.DoItSlowly(len(newPods), initialBatchSize, func() error {
		pod := <-podsCreationChan

		cs := updateGD
		if pod.Labels[apps.ControllerRevisionHashLabelKey] == currentRevision {
			cs = currentGD
		}

		var createErr error
		if createErr = r.createOnePod(cs, pod); createErr != nil {
			return createErr
		}
		created = true
		successPodNames.Store(pod.Name, struct{}{})
		return nil
	})

	// rollback to ignore failure pods because the informer won't observe these pods
	for _, pod := range newPods {
		if _, ok := successPodNames.Load(pod.Name); !ok {
			r.exp.ObserveScale(util.GetControllerKey(updateGD), expectations.Create, pod.Name)
		}
	}

	return created, err
}

func (r *realControl) createOnePod(deploy *gdv1alpha1.GameDeployment, pod *v1.Pod) error {
	startTime := time.Now()
	if _, err := r.kubeClient.CoreV1().Pods(deploy.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{}); err != nil {
		r.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedCreate", "failed to create pod: %v, pod: %v", err, util.DumpJSON(pod))
		r.metrics.CollectPodCreateDurations(util.GetControllerKey(deploy), "failure", time.Since(startTime))
		return err
	}

	r.recorder.Eventf(deploy, v1.EventTypeNormal, "SuccessfulCreate", "succeed to create pod %s", pod.Name)
	r.metrics.CollectPodCreateDurations(util.GetControllerKey(deploy), "success", time.Since(startTime))
	return nil
}

func (r *realControl) deletePods(deploy *gdv1alpha1.GameDeployment, podsToDelete []*v1.Pod, newStatus *gdv1alpha1.GameDeploymentStatus) (bool, error) {
	var deleted bool
	for _, pod := range podsToDelete {
		r.exp.ExpectScale(util.GetControllerKey(deploy), expectations.Delete, pod.Name)

		canDelete, err := r.preDeleteControl.CheckDelete(deploy, pod, newStatus, gdv1alpha1.GameDeploymentInstanceID)
		if err != nil {
			klog.V(3).Infof("preDelete check err: %s, can't delete the pod %s/%s now.", err, pod.Name, pod.Namespace)
			return deleted, err
		}
		if canDelete {
			if pod.Status.Phase != v1.PodRunning {
				klog.V(2).Infof("Pod %s/%s is not running, skip PreDelete Hook run checking.", pod.Namespace, pod.Name)
			} else if deploy.Spec.PreDeleteUpdateStrategy.Hook != nil {
				klog.V(2).Infof("PreDelete Hook run successfully, delete the pod %s/%s now.", pod.Namespace, pod.Name)
			}
		} else {
			klog.V(2).Infof("PreDelete Hook not completed, can't delete the pod %s/%s now.", pod.Namespace, pod.Name)
			continue
		}

		startTime := time.Now()
		if err := r.kubeClient.CoreV1().Pods(pod.Namespace).Delete(context.TODO(),
			pod.Name, metav1.DeleteOptions{}); err != nil {
			r.exp.ObserveScale(util.GetControllerKey(deploy), expectations.Delete, pod.Name)
			r.recorder.Eventf(deploy, v1.EventTypeWarning, "FailedDelete", "failed to delete pod %s: %v", pod.Name, err)
			r.metrics.CollectPodDeleteDurations(util.GetControllerKey(deploy), "failure", time.Since(startTime))
			return deleted, err
		}
		deleted = true
		r.recorder.Event(deploy, v1.EventTypeNormal, "SuccessfulDelete", fmt.Sprintf("succeed to delete pod %s", pod.Name))
		r.metrics.CollectPodDeleteDurations(util.GetControllerKey(deploy), "success", time.Since(startTime))
	}

	return deleted, nil
}
