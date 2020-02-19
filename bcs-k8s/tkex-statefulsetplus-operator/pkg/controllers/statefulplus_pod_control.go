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

package statefulsetplus

import (
	"fmt"
	"strings"

	stsplus "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1"

	"github.com/golang/glog"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	errorutils "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
)

// StatefulPlusPodControlInterface defines the interface that StatefulSetController uses to create, update, and delete Pods,
// and to update the Status of a StatefulSet. It follows the design paradigms used for PodControl, but its
// implementation provides for PVC creation, ordered Pod creation, ordered Pod termination, and Pod identity enforcement.
// Like controller.PodControlInterface, it is implemented as an interface to provide for testing fakes.
type StatefulPlusPodControlInterface interface {
	// CreateStatefulPlusPod create a Pod in a StatefulSet. Any PVCs necessary for the Pod are created prior to creating
	// the Pod. If the returned error is nil the Pod and its PVCs have been created.
	CreateStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error
	// UpdateStatefulPlusPod Updates a Pod in a StatefulSet. If the Pod already has the correct identity and stable
	// storage this method is a no-op. If the Pod must be mutated to conform to the Set, it is mutated and updated.
	// pod is an in-out parameter, and any updates made to the pod are reflected as mutations to this parameter. If
	// the create is successful, the returned error is nil.
	UpdateStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error
	// DeleteStatefulPlusPod deletes a Pod in a StatefulSet. The pods PVCs are not deleted. If the delete is successful,
	// the returned error is nil.
	DeleteStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error

	// ForceDeleteStatefulPlusPod force deletes a Pod in a StatefulSet. The pods PVCs are not deleted. If the delete is successful,
	// the returned error is nil.
	ForceDeleteStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error
}

func NewRealStatefulPlusPodControl(
	client clientset.Interface,
	podLister corelisters.PodLister,
	pvcLister corelisters.PersistentVolumeClaimLister,
	recorder record.EventRecorder,
) StatefulPlusPodControlInterface {
	return &realStatefulPlusPodControl{client, podLister, pvcLister, recorder}
}

// realStatefulPlusPodControl implements StatefulPlusPodControlInterface using a clientset.Interface to communicate with the
// API server. The struct is package private as the internal details are irrelevant to importing packages.
type realStatefulPlusPodControl struct {
	client    clientset.Interface
	podLister corelisters.PodLister
	pvcLister corelisters.PersistentVolumeClaimLister
	recorder  record.EventRecorder
}

func (spc *realStatefulPlusPodControl) CreateStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error {
	// Create the Pod's PVCs prior to creating the Pod
	if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
		spc.recordPodEvent("create", set, pod, err)
		return err
	}
	// If we created the PVCs attempt to create the Pod
	_, err := spc.client.CoreV1().Pods(set.Namespace).Create(pod)
	// sink already exists errors
	if apierrors.IsAlreadyExists(err) {
		return err
	}
	spc.recordPodEvent("create", set, pod, err)
	return err
}

func (spc *realStatefulPlusPodControl) UpdateStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error {
	attemptedUpdate := false
	err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		// assume the Pod is consistent
		// DeveloperJim: we support InplaceUpdate mode, so pod will not
		//   consistent except for identityMatch and storageMatch these two
		//   cases, we will update Env/image/initContainer etc. so we clean
		//   consistent condition.
		//   we had better ensure all updates are essential when calling UpdateStatefulPlusPod
		// consistent := true
		// if the Pod does not conform to its identity, update the identity and dirty the Pod
		if !IdentityMatches(set, pod) {
			updateIdentity(set, pod)
			// consistent = false
		}
		// if the Pod does not conform to the StatefulSetPlus's storage requirements, update the Pod's PVC's,
		// dirty the Pod, and create any missing PVCs
		if !storageMatches(set, pod) {
			updateStorage(set, pod)
			// consistent = false
			if err := spc.createPersistentVolumeClaims(set, pod); err != nil {
				spc.recordPodEvent("update", set, pod, err)
				return err
			}
		}
		// if the Pod is not dirty, do nothing
		// if consistent {
		// 	return nil
		// }
		attemptedUpdate = true
		// commit the update, retrying on conflicts
		_, updateErr := spc.client.CoreV1().Pods(set.Namespace).Update(pod)
		if updateErr == nil {
			glog.Infof("Pod %s/%s is updating successfully in UpdateStatefulPlusPod", pod.Namespace, pod.Name)
			return nil
		}
		glog.Errorf("Pod %s/%s update err in UpdateStatefulPlusPod: %+v", pod.Namespace, pod.Name, updateErr)
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

func (spc *realStatefulPlusPodControl) DeleteStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error {
	err := spc.client.CoreV1().Pods(set.Namespace).Delete(pod.Name, nil)
	spc.recordPodEvent("delete", set, pod, err)
	return err
}

func (spc *realStatefulPlusPodControl) ForceDeleteStatefulPlusPod(set *stsplus.StatefulSetPlus, pod *v1.Pod) error {
	err := spc.client.CoreV1().Pods(set.Namespace).Delete(pod.Name, metav1.NewDeleteOptions(0))
	spc.recordPodEvent("force delete", set, pod, err)
	return err
}

// recordPodEvent records an event for verb applied to a Pod in a StatefulSetPlus. If err is nil the generated event will
// have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a reason of v1.EventTypeWarning.
func (spc *realStatefulPlusPodControl) recordPodEvent(verb string, set *stsplus.StatefulSetPlus, pod *v1.Pod, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSetPlus %s successful",
			strings.ToLower(verb), pod.Name, set.Name)
		spc.recorder.Event(set, v1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Pod %s in StatefulSetPlus %s failed error: %s",
			strings.ToLower(verb), pod.Name, set.Name, err)
		spc.recorder.Event(set, v1.EventTypeWarning, reason, message)
	}
}

// recordClaimEvent records an event for verb applied to the PersistentVolumeClaim of a Pod in a StatefulSetPlus. If err is
// nil the generated event will have a reason of v1.EventTypeNormal. If err is not nil the generated event will have a
// reason of v1.EventTypeWarning.
func (spc *realStatefulPlusPodControl) recordClaimEvent(verb string, set *stsplus.StatefulSetPlus, pod *v1.Pod, claim *v1.PersistentVolumeClaim, err error) {
	if err == nil {
		reason := fmt.Sprintf("Successful%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s Pod %s in StatefulSetPlus %s success",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name)
		spc.recorder.Event(set, v1.EventTypeNormal, reason, message)
	} else {
		reason := fmt.Sprintf("Failed%s", strings.Title(verb))
		message := fmt.Sprintf("%s Claim %s for Pod %s in StatefulSetPlus %s failed error: %s",
			strings.ToLower(verb), claim.Name, pod.Name, set.Name, err)
		spc.recorder.Event(set, v1.EventTypeWarning, reason, message)
	}
}

// createPersistentVolumeClaims creates all of the required PersistentVolumeClaims for pod, which must be a member of
// set. If all of the claims for Pod are successfully created, the returned error is nil. If creation fails, this method
// may be called again until no error is returned, indicating the PersistentVolumeClaims for pod are consistent with
// set's Spec.
func (spc *realStatefulPlusPodControl) createPersistentVolumeClaims(set *stsplus.StatefulSetPlus, pod *v1.Pod) error {
	var errs []error
	for _, claim := range getPersistentVolumeClaims(set, pod) {
		_, err := spc.pvcLister.PersistentVolumeClaims(claim.Namespace).Get(claim.Name)
		switch {
		case apierrors.IsNotFound(err):
			_, err := spc.client.CoreV1().PersistentVolumeClaims(claim.Namespace).Create(&claim)
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

var _ StatefulPlusPodControlInterface = &realStatefulPlusPodControl{}
