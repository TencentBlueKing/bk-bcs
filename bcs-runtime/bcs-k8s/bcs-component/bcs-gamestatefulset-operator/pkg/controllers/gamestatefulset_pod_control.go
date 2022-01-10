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

package gamestatefulset

import (
	"context"
	"fmt"
	"strings"
	"time"

	stsplus "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog"
)

const podNodeLostForceDeleteKey = "pod.gamestatefulset.bkbcs.tencent.com/node-lost-force-delete"

// GameStatefulSetPodControlInterface defines the interface that StatefulSetController uses to create, update, and delete Pods,
// and to update the Status of a StatefulSet. It follows the design paradigms used for PodControl, but its
// implementation provides for PVC creation, ordered Pod creation, ordered Pod termination, and Pod identity enforcement.
// Like controller.PodControlInterface, it is implemented as an interface to provide for testing fakes.
type GameStatefulSetPodControlInterface interface {
	// CreateGameStatefulSetPod create a Pod in a StatefulSet. Any PVCs necessary for the Pod are created prior to creating
	// the Pod. If the returned error is nil the Pod and its PVCs have been created.
	CreateGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error
	// UpdateGameStatefulSetPod Updates a Pod in a StatefulSet. If the Pod already has the correct identity and stable
	// storage this method is a no-op. If the Pod must be mutated to conform to the Set, it is mutated and updated.
	// pod is an in-out parameter, and any updates made to the pod are reflected as mutations to this parameter. If
	// the create is successful, the returned error is nil.
	UpdateGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error
	// DeleteGameStatefulSetPod deletes a Pod in a StatefulSet. The pods PVCs are not deleted. If the delete is successful,
	// the returned error is nil.
	DeleteGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error

	// ForceDeleteGameStatefulSetPod force deletes a Pod in a StatefulSet. The pods PVCs are not deleted. If the delete is successful,
	// the returned error is nil.
	ForceDeleteGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) (bool, error)
}

//NewRealGameStatefulSetPodControl create implementation according GameStatefulSetPodControlInterface
func NewRealGameStatefulSetPodControl(
	client clientset.Interface,
	podLister corelisters.PodLister,
	pvcLister corelisters.PersistentVolumeClaimLister,
	nodeLister corelisters.NodeLister,
	recorder record.EventRecorder,
	metrics *metrics,
) GameStatefulSetPodControlInterface {
	return &realGameStatefulSetPodControl{client, podLister, pvcLister, nodeLister, recorder, metrics}
}

// realGameStatefulSetPodControl implements GameStatefulSetPodControlInterface using a clientset.Interface to communicate with the
// API server. The struct is package private as the internal details are irrelevant to importing packages.
type realGameStatefulSetPodControl struct {
	client     clientset.Interface
	podLister  corelisters.PodLister
	pvcLister  corelisters.PersistentVolumeClaimLister
	nodeLister corelisters.NodeLister
	recorder   record.EventRecorder
	metrics    *metrics
}

// CreateGameStatefulSetPod create pod according to definition in GameStatefulSet
func (spc *realGameStatefulSetPodControl) CreateGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error {
	// Create the Pod's PVCs prior to creating the Pod
	if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
		spc.recordPodEvent("create", set, pod, err)
		return err
	}
	// If we created the PVCs attempt to create the Pod
	startTime := time.Now()
	_, err := spc.client.CoreV1().Pods(set.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	if err == nil {
		spc.metrics.collectPodCreateDurations(set.Namespace, set.Name, "success", time.Since(startTime))
	} else {
		spc.metrics.collectPodCreateDurations(set.Namespace, set.Name, "failure", time.Since(startTime))
		klog.Infof("failed to create pod %v", pod.Name)
	}

	// sink already exists errors
	if apierrors.IsAlreadyExists(err) {
		return err
	}
	spc.recordPodEvent("create", set, pod, err)
	return err
}

// UpdateGameStatefulSetPod update pod info of GameStatefulSet
func (spc *realGameStatefulSetPodControl) UpdateGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error {
	attemptedUpdate := false
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// assume the Pod is consistent
		consistent := true
		// if the Pod does not conform to its identity, update the identity and dirty the Pod
		if !IdentityMatches(set, pod) {
			updateIdentity(set, pod)
			consistent = false
		}
		// if the Pod does not conform to the GameStatefulSet's storage requirements, update the Pod's PVC's,
		// dirty the Pod, and create any missing PVCs
		if !storageMatches(set, pod) {
			updateStorage(set, pod)
			consistent = false
			if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
				spc.recordPodEvent("update", set, pod, err)
				return err
			}
		}
		// if the Pod is not dirty, do nothing
		if consistent {
			return nil
		}
		attemptedUpdate = true
		// commit the update, retrying on conflicts
		_, updateErr := spc.client.CoreV1().Pods(set.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
		if updateErr == nil {
			klog.Infof("Pod %s/%s is updating successfully in UpdateGameStatefulSetPod", pod.Namespace, pod.Name)
			return nil
		}
		klog.Errorf("Pod %s/%s update err in UpdateGameStatefulSetPod: %+v", pod.Namespace, pod.Name, updateErr)
		if updated, err := spc.podLister.Pods(set.Namespace).Get(pod.Name); err == nil {
			// make a copy so we don't mutate the shared cache
			pod = updated.DeepCopy()
		} else {
			utilruntime.HandleError(fmt.Errorf("error getting updated Pod %s/%s from lister: %v", set.Namespace, pod.Name, err))
		}

		return updateErr
	})
	if attemptedUpdate {
		spc.recordPodEvent("update", set, pod, err)
	}
	return err
}

// DeleteGameStatefulSetPod delete pod according to GameStatefulSet
func (spc *realGameStatefulSetPodControl) DeleteGameStatefulSetPod(set *stsplus.GameStatefulSet, pod *v1.Pod) error {
	startTime := time.Now()
	err := spc.client.CoreV1().Pods(set.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
	if err == nil {
		spc.metrics.collectPodDeleteDurations(set.Namespace, set.Name, "success", time.Since(startTime))
	} else {
		spc.metrics.collectPodDeleteDurations(set.Namespace, set.Name, "failure", time.Since(startTime))
	}
	spc.recordPodEvent("delete", set, pod, err)
	return err
}

// ForceDeleteGameStatefulSetPod delete pod according to GameStatefulSet
func (spc *realGameStatefulSetPodControl) ForceDeleteGameStatefulSetPod(set *stsplus.GameStatefulSet,
	pod *v1.Pod) (bool, error) {
	v, ok := pod.Annotations[podNodeLostForceDeleteKey]
	if !ok || v != "true" {
		klog.Infof("GameStatefulSet %s/%s's Pod %s/%s need not to be force deleted, "+
			"because annotation is not set", set.Namespace, set.Name, pod.Namespace, pod.Name)
		return false, nil
	}

	node, errNode := spc.nodeLister.Get(pod.Spec.NodeName)
	if errNode != nil {
		return false, errNode
	}

	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == v1.NodeReady {
			if node.Status.Conditions[i].Status != v1.ConditionTrue {
				klog.Infof("GameStatefulSet %s/%s's Pod %s/%s need to be force deleted", set.Namespace,
					set.Name, pod.Namespace, pod.Name)
				break
			} else {
				klog.Infof("GameStatefulSet %s/%s's Pod %s/%s need not to be force deleted", set.Namespace,
					set.Name, pod.Namespace, pod.Name)
				return false, nil
			}
		}
	}

	err := spc.client.CoreV1().Pods(set.Namespace).Delete(context.TODO(), pod.Name, *metav1.NewDeleteOptions(0))
	if err != nil {
		klog.Errorf("GameStatefulSet %s/%s's Pod %s/%s force delete error", set.Namespace, set.Name,
			pod.Namespace, pod.Name)
		return false, err
	}

	spc.recordPodEvent("force delete", set, pod, err)
	return true, err
}

// recordPodEvent records an event for verb applied to a Pod in a GameStatefulSet. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (spc *realGameStatefulSetPodControl) recordPodEvent(verb string, set *stsplus.GameStatefulSet, pod *v1.Pod, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in GameStatefulSet %s successful",
			strings.ToLower(verb), pod.Name, set.Name)
		spc.recorder.Event(set, v1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in GameStatefulSet %s failed error: %s",
			strings.ToLower(verb), pod.Name, set.Name, err)
		spc.recorder.Event(set, v1.EventTypeWarning, reason, message)
	}
}

// recordClaimEvent records an event for verb applied to the PersistentVolumeClaim of a Pod in a GameStatefulSet. If err is
// nil the generated event will have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a
// reason of v1.EventTypeWarning.
func (spc *realGameStatefulSetPodControl) recordClaimEvent(verb string, set *stsplus.GameStatefulSet, pod *v1.Pod, claim *v1.PersistentVolumeClaim, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s Pod %s in GameStatefulSet %s success",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name)
		spc.recorder.Event(set, v1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s for Pod %s in GameStatefulSet %s failed error: %s",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name, err)
		spc.recorder.Event(set, v1.EventTypeWarning, reason, message)
	}
}

// createPersistentVolumeClaims creates all of the required PersistentVolumeClaims for pod, which must be a member of
// set. If all of the claims for Pod are successfully created, the returned error is nil. If creation fails, this method
// may be called again until no error is returned, indicating the PersistentVolumeClaims for pod are consistent with
// set's Spec.
func (spc *realGameStatefulSetPodControl) createPersistentVolumeClaims(set *stsplus.GameStatefulSet, pod *v1.Pod) error {
	var errs []error
	for _, claim := range getPersistentVolumeClaims(set, pod) {
		_, err := spc.pvcLister.PersistentVolumeClaims(claim.Namespace).Get(claim.Name)
		switch {
		case apierrors.IsNotFound(err):
			_, err := spc.client.CoreV1().PersistentVolumeClaims(claim.Namespace).Create(context.TODO(),
				&claim, metav1.CreateOptions{})
			if err != nil {
				errs = append(errs, fmt.Errorf("Failed to create PVC %s: %s", claim.Name, err))
			}
			if err == nil || !apierrors.IsAlreadyExists(err) {
				spc.recordClaimEvent("create", set, pod, &claim, err)
			}
		case err != nil:
			errs = append(errs, fmt.Errorf("Failed to retrieve PVC %s: %s", claim.Name, err))
			spc.recordClaimEvent("create", set, pod, &claim, err)
		}
		// TODO: Check resource requirements and accessmodes, update if necessary
	}
	return errorutils.NewAggregate(errs)
}

var _ GameStatefulSetPodControlInterface = &realGameStatefulSetPodControl{}
