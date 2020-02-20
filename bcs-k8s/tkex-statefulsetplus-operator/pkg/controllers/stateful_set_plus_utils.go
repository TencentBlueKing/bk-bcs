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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	stsplus "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/apis/tkex/v1alpha1"
	stspluslisters "bk-bcs/bcs-k8s/tkex-statefulsetplus-operator/pkg/listers/tkex/v1alpha1"

	"github.com/golang/glog"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes/scheme"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	node "k8s.io/kubernetes/pkg/util/node"
)

// maxUpdateRetries is the maximum number of retries used for update conflict resolution prior to failure
const maxUpdateRetries = 10

// updateConflictError is the error used to indicate that the maximum number of retries against the API server have
// been attempted and we need to back off
var updateConflictError = fmt.Errorf("aborting update after %d attempts", maxUpdateRetries)
var patchCodec = scheme.Codecs.LegacyCodec(stsplus.SchemeGroupVersion)

// overlappingStatefulSetPluses sorts a list of StatefulSetPluses by creation timestamp, using their names as a tie breaker.
// Generally used to tie break between StatefulSetPluses that have overlapping selectors.
type overlappingStatefulSetPluses []*stsplus.StatefulSetPlus

func (o overlappingStatefulSetPluses) Len() int { return len(o) }

func (o overlappingStatefulSetPluses) Swap(i, j int) { o[i], o[j] = o[j], o[i] }

func (o overlappingStatefulSetPluses) Less(i, j int) bool {
	if o[i].CreationTimestamp.Equal(&o[j].CreationTimestamp) {
		return o[i].Name < o[j].Name
	}
	return o[i].CreationTimestamp.Before(&o[j].CreationTimestamp)
}

// statefulPodRegex is a regular expression that extracts the parent StatefulSetPlus and ordinal from the Name of a Pod
var statefulPlusPodRegex = regexp.MustCompile("(.*)-([0-9]+)$")

// getParentNameAndOrdinal gets the name of pod's parent StatefulSetPlus and pod's ordinal as extracted from its Name. If
// the Pod was not created by a StatefulSetPlus, its parent is considered to be empty string, and its ordinal is considered
// to be -1.
func getParentNameAndOrdinal(pod *v1.Pod) (string, int) {
	parent := ""
	ordinal := -1
	subMatches := statefulPlusPodRegex.FindStringSubmatch(pod.Name)
	if len(subMatches) < 3 {
		return parent, ordinal
	}
	parent = subMatches[1]
	if i, err := strconv.ParseInt(subMatches[2], 10, 32); err == nil {
		ordinal = int(i)
	}
	return parent, ordinal
}

// getParentName gets the name of pod's parent StatefulSetPlus. If pod has not parent, the empty string is returned.
func getParentName(pod *v1.Pod) string {
	parent, _ := getParentNameAndOrdinal(pod)
	return parent
}

//  getOrdinal gets pod's ordinal. If pod has no ordinal, -1 is returned.
func getOrdinal(pod *v1.Pod) int {
	_, ordinal := getParentNameAndOrdinal(pod)
	return ordinal
}

// getPodName gets the name of set's child Pod with an ordinal index of ordinal
func getPodName(set *stsplus.StatefulSetPlus, ordinal int) string {
	return fmt.Sprintf("%s-%d", set.Name, ordinal)
}

// getPersistentVolumeClaimName gets the name of PersistentVolumeClaim for a Pod with an ordinal index of ordinal. claim
// must be a PersistentVolumeClaim from set's VolumeClaims template.
func getPersistentVolumeClaimName(set *stsplus.StatefulSetPlus, claim *v1.PersistentVolumeClaim, ordinal int) string {
	// NOTE: This name format is used by the heuristics for zone spreading in ChooseZoneForVolume
	return fmt.Sprintf("%s-%s-%d", claim.Name, set.Name, ordinal)
}

// isMemberOf tests if pod is a member of set.
func isMemberOf(set *stsplus.StatefulSetPlus, pod *v1.Pod) bool {
	return getParentName(pod) == set.Name
}

// identityMatches returns true if pod has a valid identity and network identity for a member of set.
func IdentityMatches(set *stsplus.StatefulSetPlus, pod *v1.Pod) bool {
	parent, ordinal := getParentNameAndOrdinal(pod)
	return ordinal >= 0 &&
		set.Name == parent &&
		pod.Name == getPodName(set, ordinal) &&
		pod.Namespace == set.Namespace &&
		pod.Labels[stsplus.StatefulSetPlusPodNameLabel] == pod.Name
}

// storageMatches returns true if pod's Volumes cover the set of PersistentVolumeClaims
func storageMatches(set *stsplus.StatefulSetPlus, pod *v1.Pod) bool {
	ordinal := getOrdinal(pod)
	if ordinal < 0 {
		return false
	}
	volumes := make(map[string]v1.Volume, len(pod.Spec.Volumes))
	for _, volume := range pod.Spec.Volumes {
		volumes[volume.Name] = volume
	}
	for _, claim := range set.Spec.VolumeClaimTemplates {
		volume, found := volumes[claim.Name]
		if !found ||
			volume.VolumeSource.PersistentVolumeClaim == nil ||
			volume.VolumeSource.PersistentVolumeClaim.ClaimName !=
				getPersistentVolumeClaimName(set, &claim, ordinal) {
			return false
		}
	}
	return true
}

// getPersistentVolumeClaims gets a map of PersistentVolumeClaims to their template names, as defined in set. The
// returned PersistentVolumeClaims are each constructed with a the name specific to the Pod. This name is determined
// by getPersistentVolumeClaimName.
func getPersistentVolumeClaims(set *stsplus.StatefulSetPlus, pod *v1.Pod) map[string]v1.PersistentVolumeClaim {
	ordinal := getOrdinal(pod)
	templates := set.Spec.VolumeClaimTemplates
	claims := make(map[string]v1.PersistentVolumeClaim, len(templates))
	for i := range templates {
		claim := templates[i]
		claim.Name = getPersistentVolumeClaimName(set, &claim, ordinal)
		claim.Namespace = set.Namespace
		if claim.Labels != nil {
			for key, value := range set.Spec.Selector.MatchLabels {
				claim.Labels[key] = value
			}
		} else {
			claim.Labels = set.Spec.Selector.MatchLabels
		}
		claims[templates[i].Name] = claim
	}
	return claims
}

// updateStorage updates pod's Volumes to conform with the PersistentVolumeClaim of set's templates. If pod has
// conflicting local Volumes these are replaced with Volumes that conform to the set's templates.
func updateStorage(set *stsplus.StatefulSetPlus, pod *v1.Pod) {
	currentVolumes := pod.Spec.Volumes
	claims := getPersistentVolumeClaims(set, pod)
	newVolumes := make([]v1.Volume, 0, len(claims))
	for name, claim := range claims {
		newVolumes = append(newVolumes, v1.Volume{
			Name: name,
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: claim.Name,
					// TODO: Use source definition to set this value when we have one.
					ReadOnly: false,
				},
			},
		})
	}
	for i := range currentVolumes {
		if _, ok := claims[currentVolumes[i].Name]; !ok {
			newVolumes = append(newVolumes, currentVolumes[i])
		}
	}
	pod.Spec.Volumes = newVolumes
}

func initIdentity(set *stsplus.StatefulSetPlus, pod *v1.Pod) {
	updateIdentity(set, pod)
	// Set these immutable fields only on initial Pod creation, not updates.
	pod.Spec.Hostname = pod.Name
	pod.Spec.Subdomain = set.Spec.ServiceName
}

// updateIdentity updates pod's name, hostname, and subdomain, and StatefulSetPlusPodNameLabel to conform to set's name
// and headless service.
func updateIdentity(set *stsplus.StatefulSetPlus, pod *v1.Pod) {
	pod.Name = getPodName(set, getOrdinal(pod))
	pod.Namespace = set.Namespace
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[stsplus.StatefulSetPlusPodNameLabel] = pod.Name
}

// isRunningAndReady returns true if pod is in the PodRunning Phase, if it has a condition of PodReady.
func isRunningAndReady(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning && podutil.IsPodReady(pod)
}

// isCreated returns true if pod has been created and is maintained by the API server
func isCreated(pod *v1.Pod) bool {
	return pod.Status.Phase != ""
}

// isFailed returns true if pod has a Phase of PodFailed
func isFailed(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodFailed
}

// isOwnedNodeLost return true if pod has a Reason of NodeLost
func isOwnedNodeLost(pod *v1.Pod) bool {
	return pod.Status.Reason == node.NodeUnreachablePodReason
}

// isTerminating returns true if pod's DeletionTimestamp has been set
func isTerminating(pod *v1.Pod) bool {
	return pod.DeletionTimestamp != nil
}

// isHealthy returns true if pod is running and ready and has not been terminated
func isHealthy(pod *v1.Pod) bool {
	return isRunningAndReady(pod) && !isTerminating(pod)
}

// allowsBurst is true if the alpha burst annotation is set.
func allowsBurst(set *stsplus.StatefulSetPlus) bool {
	return set.Spec.PodManagementPolicy == stsplus.ParallelPodManagement
}

// setPodRevision sets the revision of Pod to revision by adding the StatefulSetPlusRevisionLabel
func setPodRevision(pod *v1.Pod, revision string) {
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[stsplus.StatefulSetPlusRevisionLabel] = revision
}

// getPodRevision gets the revision of Pod by inspecting the StatefulSetPlusRevisionLabel. If pod has no revision the empty
// string is returned.
func getPodRevision(pod *v1.Pod) string {
	if pod.Labels == nil {
		return ""
	}
	return pod.Labels[stsplus.StatefulSetPlusRevisionLabel]
}

// newStatefulSetPlusPod returns a new Pod conforming to the set's Spec with an identity generated from ordinal.
func newStatefulSetPlusPod(set *stsplus.StatefulSetPlus, ordinal int) *v1.Pod {
	pod, _ := controller.GetPodFromTemplate(&set.Spec.Template, set, metav1.NewControllerRef(set, controllerKind))
	pod.Name = getPodName(set, ordinal)
	initIdentity(set, pod)
	updateStorage(set, pod)
	return pod
}

// newVersionedStatefulSetPlusPod creates a new Pod for a StatefulSetPlus. currentSet is the representation of the set at the
// current revision. updateSet is the representation of the set at the updateRevision. currentRevision is the name of
// the current revision. updateRevision is the name of the update revision. ordinal is the ordinal of the Pod. If the
// returned error is nil, the returned Pod is valid.
func newVersionedStatefulSetPlusPod(currentSet, updateSet *stsplus.StatefulSetPlus, currentRevision, updateRevision string, ordinal int) *v1.Pod {
	if currentSet.Spec.UpdateStrategy.Type != stsplus.OnDeleteStatefulSetPlusStrategyType &&
		(currentSet.Spec.UpdateStrategy.RollingUpdate == nil && ordinal < int(currentSet.Status.CurrentReplicas)) ||
		(currentSet.Spec.UpdateStrategy.RollingUpdate != nil && ordinal < int(*currentSet.Spec.UpdateStrategy.RollingUpdate.Partition)) {
		pod := newStatefulSetPlusPod(currentSet, ordinal)
		setPodRevision(pod, currentRevision)
		return pod
	}
	pod := newStatefulSetPlusPod(updateSet, ordinal)
	setPodRevision(pod, updateRevision)
	return pod
}

// updateVersionedStatefulSetPlusPod creates a new Pod for a StatefulSetPlus. currentSet is the representation of the set at the
// current revision. updateSet is the representation of the set at the updateRevision. currentRevision is the name of
// the current revision. updateRevision is the name of the update revision. ordinal is the ordinal of the Pod. If the
// returned error is nil, the returned Pod is valid.
func updateVersionedStatefulSetPlusPod(updateSet *stsplus.StatefulSetPlus, updateRevision string, oldPod *v1.Pod) *v1.Pod {
	// Make a deep copy so we don't mutate the shared cache
	pod := oldPod.DeepCopy()
	//copy Pod Labels & Annotations
	if pod.Labels == nil {
		pod.Labels = make(labels.Set)
	}
	for k, v := range updateSet.Spec.Template.Labels {
		pod.Labels[k] = v
	}
	if pod.Annotations == nil {
		pod.Annotations = make(labels.Set)
	}
	for k, v := range updateSet.Spec.Template.Annotations {
		pod.Annotations[k] = v
	}
	if updateSet.Spec.Template.Spec.ActiveDeadlineSeconds != nil {
		pod.Spec.ActiveDeadlineSeconds = updateSet.Spec.Template.Spec.ActiveDeadlineSeconds
	}

	if len(pod.Spec.InitContainers) > 0 {
		// Deep copy initContainers
		desiredInitContainers := make(map[string]v1.Container)
		for _, container := range updateSet.Spec.Template.Spec.InitContainers {
			desiredInitContainers[container.Name] = container
		}
		for i := range pod.Spec.InitContainers {
			if container, ok := desiredInitContainers[pod.Spec.InitContainers[i].Name]; ok {
				pod.Spec.InitContainers[i].Image = container.Image
			}
		}
	}

	// Deep copy container
	desiredContainers := make(map[string]v1.Container)
	for _, container := range updateSet.Spec.Template.Spec.Containers {
		desiredContainers[container.Name] = container
	}
	for i := range pod.Spec.Containers {
		if container, ok := desiredContainers[pod.Spec.Containers[i].Name]; ok {
			pod.Spec.Containers[i].Image = container.Image
		}
	}

	// TODO
	// only additions to existing tolerations

	setPodRevision(pod, updateRevision)
	return pod
}

// Match check if the given StatefulSetPlus's template matches the template stored in the given history.
func Match(ss *stsplus.StatefulSetPlus, history *apps.ControllerRevision) (bool, error) {
	patch, err := getPatch(ss)
	if err != nil {
		return false, err
	}
	return bytes.Equal(patch, history.Data.Raw), nil
}

// getPatch returns a strategic merge patch that can be applied to restore a StatefulSetPlus to a
// previous version. If the returned error is nil the patch is valid. The current state that we save is just the
// PodSpecTemplate. We can modify this later to encompass more state (or less) and remain compatible with previously
// recorded patches.
func getPatch(set *stsplus.StatefulSetPlus) ([]byte, error) {
	str, err := runtime.Encode(patchCodec, set)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	json.Unmarshal([]byte(str), &raw)
	objCopy := make(map[string]interface{})
	specCopy := make(map[string]interface{})
	spec := raw["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	specCopy["template"] = template
	template["$patch"] = "replace"
	objCopy["spec"] = specCopy
	patch, err := json.Marshal(objCopy)
	return patch, err
}

// newRevision creates a new ControllerRevision containing a patch that reapplies the target state of set.
// The Revision of the returned ControllerRevision is set to revision. If the returned error is nil, the returned
// ControllerRevision is valid. StatefulSetPlus revisions are stored as patches that re-apply the current state of set
// to a new StatefulSetPlus using a strategic merge patch to replace the saved state of the new StatefulSetPlus.
func newRevision(set *stsplus.StatefulSetPlus, revision int64, collisionCount *int32) (*apps.ControllerRevision, error) {
	patch, err := getPatch(set)
	if err != nil {
		return nil, err
	}
	cr, err := history.NewControllerRevision(set,
		controllerKind,
		set.Spec.Template.Labels,
		runtime.RawExtension{Raw: patch},
		revision,
		collisionCount)
	if err != nil {
		return nil, err
	}
	if cr.ObjectMeta.Annotations == nil {
		cr.ObjectMeta.Annotations = make(map[string]string)
	}
	for key, value := range set.Annotations {
		cr.ObjectMeta.Annotations[key] = value
	}
	return cr, nil
}

// ApplyRevision returns a new StatefulSetPlus constructed by restoring the state in revision to set. If the returned error
// is nil, the returned StatefulSetPlus is valid.
func ApplyRevision(set *stsplus.StatefulSetPlus, revision *apps.ControllerRevision) (*stsplus.StatefulSetPlus, error) {
	clone := set.DeepCopy()
	patched, err := strategicpatch.StrategicMergePatch([]byte(runtime.EncodeOrDie(patchCodec, clone)), revision.Data.Raw, clone)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(patched, clone)
	if err != nil {
		return nil, err
	}
	return clone, nil
}

// nextRevision finds the next valid revision number based on revisions. If the length of revisions
// is 0 this is 1. Otherwise, it is 1 greater than the largest revision's Revision. This method
// assumes that revisions has been sorted by Revision.
func nextRevision(revisions []*apps.ControllerRevision) int64 {
	count := len(revisions)
	if count <= 0 {
		return 1
	}
	return revisions[count-1].Revision + 1
}

// inconsistentStatus returns true if the ObservedGeneration of status is greater than set's
// Generation or if any of the status's fields do not match those of set's status.
func inconsistentStatus(set *stsplus.StatefulSetPlus, status *stsplus.StatefulSetPlusStatus) bool {
	return status.ObservedGeneration > set.Status.ObservedGeneration ||
		status.Replicas != set.Status.Replicas ||
		status.CurrentReplicas != set.Status.CurrentReplicas ||
		status.ReadyReplicas != set.Status.ReadyReplicas ||
		status.UpdatedReplicas != set.Status.UpdatedReplicas ||
		status.CurrentRevision != set.Status.CurrentRevision ||
		status.UpdateRevision != set.Status.UpdateRevision
}

// completeRollingUpdate completes a rolling update when all of set's replica Pods have been updated
// to the updateRevision. status's currentRevision is set to updateRevision and its' updateRevision
// is set to the empty string. status's currentReplicas is set to updateReplicas and its updateReplicas
// are set to 0.
func completeRollingUpdate(set *stsplus.StatefulSetPlus, status *stsplus.StatefulSetPlusStatus) {
	if set.Spec.UpdateStrategy.Type != stsplus.OnDeleteStatefulSetPlusStrategyType &&
		status.UpdatedReplicas == status.Replicas &&
		status.ReadyReplicas == status.Replicas {
		status.CurrentReplicas = status.UpdatedReplicas
		status.CurrentRevision = status.UpdateRevision
	}
}

// ascendingOrdinal is a sort.Interface that Sorts a list of Pods based on the ordinals extracted
// from the Pod. Pod's that have not been constructed by StatefulSetPlus's have an ordinal of -1, and are therefore pushed
// to the front of the list.
type ascendingOrdinal []*v1.Pod

func (ao ascendingOrdinal) Len() int {
	return len(ao)
}

func (ao ascendingOrdinal) Swap(i, j int) {
	ao[i], ao[j] = ao[j], ao[i]
}

func (ao ascendingOrdinal) Less(i, j int) bool {
	return getOrdinal(ao[i]) < getOrdinal(ao[j])
}

// GetPodStatefulSets returns a list of StatefulSets that potentially match a pod.
// Only the one specified in the Pod's ControllerRef will actually manage it.
// Returns an error only if no matching StatefulSets are found.
func GetPodStatefulSetPluses(pod *v1.Pod, sscLister stspluslisters.StatefulSetPlusLister) ([]*stsplus.StatefulSetPlus, error) {
	var selector labels.Selector
	var ps *stsplus.StatefulSetPlus

	if len(pod.Labels) == 0 {
		return nil, fmt.Errorf("no StatefulSets found for pod %v because it has no labels", pod.Name)
	}

	list, err := sscLister.StatefulSetPluses(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var psList []*stsplus.StatefulSetPlus
	for i := range list {
		ps = list[i]
		if ps.Namespace != pod.Namespace {
			continue
		}
		selector, err = metav1.LabelSelectorAsSelector(ps.Spec.Selector)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %v", err)
		}

		// If a StatefulSet with a nil or empty selector creeps in, it should match nothing, not everything.
		if selector.Empty() || !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		psList = append(psList, ps)
	}

	if len(psList) == 0 {
		return nil, fmt.Errorf("could not find StatefulSet for pod %s in namespace %s with labels: %v", pod.Name, pod.Namespace, pod.Labels)
	}

	return psList, nil
}

func isOnDeleteUpdateStragtegy(set *stsplus.StatefulSetPlus) bool {
	if set == nil {
		glog.Errorf("the input statefulsetplus of isOnDeleteUpdateStragtegy is nil, please check it.")
		return false
	}

	if set.Spec.UpdateStrategy.Type != stsplus.OnDeleteStatefulSetPlusStrategyType {
		glog.Errorf("the statefulsetplus's UpdateStrategy is %s, not OnDelete.", set.Spec.UpdateStrategy)
		return false
	}

	return true
}

func Contain(obj interface{}, target interface{}) (bool, error) {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true, nil
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true, nil
		}
	}

	return false, errors.New("not in array")
}

//isInplaceUpdate check if StatefulSetPlus is in InplaceUpdate mode,
func isInplaceUpdate(set *stsplus.StatefulSetPlus) bool {
	return set.Spec.UpdateStrategy.Type == stsplus.InplaceUpdateStatefulSetPlusStrategyType &&
		set.Status.CurrentRevision != set.Status.UpdateRevision
}
