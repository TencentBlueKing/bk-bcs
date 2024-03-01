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

package simulator

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/kubernetes/pkg/kubelet/types"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-cluster-autoscaler/util"
)

const (
	// PodDeletionTimeout - time after which a pod to be deleted is not included in the list of pods for drain.
	PodDeletionTimeout = 12 * time.Minute
	// PodLongTerminatingExtraThreshold - time after which a pod, that is terminating and that has run over its
	// terminationGracePeriod, should be ignored and considered as deleted
	PodLongTerminatingExtraThreshold = 30 * time.Second
)

const (
	// PodSafeToEvictKey - annotation that ignores constraints to evict a pod like not being replicated, being on
	// kube-system namespace or having a local storage.
	PodSafeToEvictKey = "cluster-autoscaler.kubernetes.io/safe-to-evict"
)

// BlockingPod represents a pod which is blocking the scale down of a node.
type BlockingPod struct {
	Pod    *apiv1.Pod
	Reason BlockingPodReason
}

// BlockingPodReason represents a reason why a pod is blocking the scale down of a node.
type BlockingPodReason string

const (
	// BlockingNoReason xxx
	// NoReason - sanity check, this should never be set explicitly. If this is found in the wild, it means that it was
	// implicitly initialized and might indicate a bug.
	BlockingNoReason BlockingPodReason = "BlockingNoReason"
	// ControllerNotFound - pod is blocking scale down because its controller can't be found.
	ControllerNotFound BlockingPodReason = "ControllerNotFound"
	// MinReplicasReached - pod is blocking scale down because its controller already has the minimum number of replicas.
	MinReplicasReached BlockingPodReason = "MinReplicasReached"
	// NotReplicated - pod is blocking scale down because it's not replicated.
	NotReplicated BlockingPodReason = "NotReplicated"
	// LocalStorageRequested - pod is blocking scale down because it requests local storage.
	LocalStorageRequested BlockingPodReason = "LocalStorageRequested"
	// NotSafeToEvictAnnotation - pod is blocking scale down because it has a "not safe to evict" annotation.
	NotSafeToEvictAnnotation BlockingPodReason = "NotSafeToEvictAnnotation"
	// UnmovableKubeSystemPod - pod is blocking scale down because it's a non-daemonset, non-mirrored,
	// non-pdb-assigned kube-system pod.
	UnmovableKubeSystemPod BlockingPodReason = "UnmovableKubeSystemPod"
	// NotEnoughPdb - pod is blocking scale down because it doesn't have enough PDB left.
	NotEnoughPdb BlockingPodReason = "NotEnoughPdb"
	// BlockingUnexpectedError - pod is blocking scale down because of an unexpected error.
	BlockingUnexpectedError BlockingPodReason = "BlockingUnexpectedError"
)

// GetPodsForDeletionOnNodeDrain returns pods that should be deleted on node drain as well as some extra information
// about possibly problematic pods (unreplicated and daemonsets).
func GetPodsForDeletionOnNodeDrain(
	podList []*apiv1.Pod,
	pdbs []*policyv1.PodDisruptionBudget,
	deleteAll bool,
	skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool,
	checkReferences bool, // Setting this to true requires client to be not-null.
	listers kube_util.ListerRegistry,
	minReplica int32,
	currentTime time.Time) (pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod, blockingPod *BlockingPod, err error) {

	pods = []*apiv1.Pod{}
	daemonSetPods = []*apiv1.Pod{}
	// filter kube-system PDBs to avoid doing it for every kube-system pod
	kubeSystemPDBs := make([]*policyv1.PodDisruptionBudget, 0)
	for _, pdb := range pdbs {
		if pdb.Namespace == "kube-system" {
			kubeSystemPDBs = append(kubeSystemPDBs, pdb)
		}
	}

	for _, pod := range podList {
		if IsMirrorPod(pod) {
			continue
		}

		// Possibly skip a pod under deletion but only if it was being deleted for long enough
		// to avoid a situation when we delete the empty node immediately after the pod was marked for
		// deletion without respecting any graceful termination.
		if IsPodLongTerminating(pod, currentTime) {
			// pod is being deleted for long enough - no need to care about it.
			continue
		}

		// daemonsetPod := false
		// replicated := false
		safeToEvict := hasSafeToEvictAnnotation(pod)
		terminal := isPodTerminal(pod)

		controllerRef := ControllerRef(pod)
		refKind := ""
		if controllerRef != nil {
			refKind = controllerRef.Kind
		}

		daemonsetPod, replicated, blockingPod, err := checkControllerRef(pod, controllerRef,
			refKind, checkReferences, minReplica, listers)
		if err != nil {
			return []*apiv1.Pod{}, []*apiv1.Pod{}, blockingPod, err
		}

		if daemonsetPod {
			daemonSetPods = append(daemonSetPods, pod)
			continue
		}

		if !deleteAll && !safeToEvict && !terminal {
			if !replicated {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: NotReplicated},
					fmt.Errorf("%s/%s is not replicated", pod.Namespace, pod.Name)
			}
			if pod.Namespace == "kube-system" && skipNodesWithSystemPods {
				hasPDB, err := checkKubeSystemPDBs(pod, kubeSystemPDBs)
				if err != nil {
					return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: BlockingUnexpectedError},
						fmt.Errorf("error matching pods to pdbs: %v", err)
				}
				if !hasPDB {
					return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: UnmovableKubeSystemPod},
						fmt.Errorf("non-daemonset, non-mirrored, non-pdb-assigned kube-system pod present: %s", pod.Name)
				}
			}
			if HasLocalStorage(pod) && skipNodesWithLocalStorage {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: LocalStorageRequested},
					fmt.Errorf("pod with local storage present: %s", pod.Name)
			}
			if hasNotSafeToEvictAnnotation(pod) {
				return []*apiv1.Pod{}, []*apiv1.Pod{}, &BlockingPod{Pod: pod, Reason: NotSafeToEvictAnnotation},
					fmt.Errorf("pod annotated as not safe to evict present: %s", pod.Name)
			}
		}
		pods = append(pods, pod)
	}
	return pods, daemonSetPods, nil, nil
}

func checkControllerRef(pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	refKind string, checkReferences bool, minReplica int32,
	listers kube_util.ListerRegistry,
) (bool, bool, *BlockingPod, error) {
	daemonsetPod := false
	replicated := false

	// For now, owner controller must be in the same namespace as the pod
	// so OwnerReference doesn't have its own Namespace field
	// controllerNamespace := pod.Namespace
	// nolint
	if refKind == "ReplicationController" {
		blockingPod, err := checkRC(&replicated, checkReferences, pod, controllerRef, listers, minReplica)
		if err != nil {
			return daemonsetPod, replicated, blockingPod, err
		}
	} else if refKind == "DaemonSet" {
		daemonsetPod = true
		blockingPod, err := checkDS(&replicated, checkReferences, pod, controllerRef, listers, minReplica)
		if err != nil {
			return daemonsetPod, replicated, blockingPod, err
		}
	} else if refKind == "Job" {
		blockingPod, err := checkJob(&replicated, checkReferences, pod, controllerRef, listers, minReplica)
		if err != nil {
			return daemonsetPod, replicated, blockingPod, err
		}
	} else if refKind == "ReplicaSet" {
		blockingPod, err := checkRS(&replicated, checkReferences, pod, controllerRef, listers, minReplica)
		if err != nil {
			return daemonsetPod, replicated, blockingPod, err
		}
	} else if refKind == "StatefulSet" {
		blockingPod, err := checkSts(&replicated, checkReferences, pod, controllerRef, listers, minReplica)
		if err != nil {
			return daemonsetPod, replicated, blockingPod, err
		}
	}
	return daemonsetPod, replicated, &BlockingPod{}, nil
}

// ControllerRef returns the OwnerReference to pod's controller.
func ControllerRef(pod *apiv1.Pod) *metav1.OwnerReference {
	return metav1.GetControllerOf(pod)
}

// IsMirrorPod checks whether the pod is a mirror pod.
func IsMirrorPod(pod *apiv1.Pod) bool {
	_, found := pod.ObjectMeta.Annotations[types.ConfigMirrorAnnotationKey]
	return found
}

// isPodTerminal checks whether the pod is in a terminal state.
func isPodTerminal(pod *apiv1.Pod) bool {
	// pod will never be restarted
	if pod.Spec.RestartPolicy == apiv1.RestartPolicyNever && (pod.Status.Phase == apiv1.PodSucceeded ||
		pod.Status.Phase == apiv1.PodFailed) {
		return true
	}
	// pod has run to completion and succeeded
	if pod.Spec.RestartPolicy == apiv1.RestartPolicyOnFailure && pod.Status.Phase == apiv1.PodSucceeded {
		return true
	}
	// kubelet has rejected this pod, due to eviction or some other constraint
	return pod.Status.Phase == apiv1.PodFailed
}

// HasLocalStorage returns true if pod has any local storage.
func HasLocalStorage(pod *apiv1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if isLocalVolume(&volume) {
			return true
		}
	}
	return false
}

func isLocalVolume(volume *apiv1.Volume) bool {
	return volume.HostPath != nil || volume.EmptyDir != nil
}

// checkKubeSystemPDBs xxx
// This only checks if a matching PDB exist and therefore if it makes sense to attempt drain simulation,
// as we check for allowed-disruptions later anyway (for all pods with PDB, not just in kube-system)
func checkKubeSystemPDBs(pod *apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (bool, error) {
	for _, pdb := range pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return false, err
		}
		if selector.Matches(labels.Set(pod.Labels)) {
			return true, nil
		}
	}

	return false, nil
}

// hasSafeToEvictAnnotation xxx
// This checks if pod has PodSafeToEvictKey annotation
func hasSafeToEvictAnnotation(pod *apiv1.Pod) bool {
	return pod.GetAnnotations()[PodSafeToEvictKey] == "true"
}

// hasNotSafeToEvictAnnotation xxx
// This checks if pod has PodSafeToEvictKey annotation set to false
func hasNotSafeToEvictAnnotation(pod *apiv1.Pod) bool {
	return pod.GetAnnotations()[PodSafeToEvictKey] == "false"
}

// FastGetPodsToMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod.
// Based on kubectl drain code. It makes an assumption that RC, DS, Jobs and RS were deleted
// along with their pods (no abandoned pods with dangling created-by annotation). Useful for fast
// checks.
func FastGetPodsToMove(nodeInfo *schedulernodeinfo.NodeInfo, skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool, pdbs []*policyv1.PodDisruptionBudget) (pods []*apiv1.Pod,
	daemonSetPods []*apiv1.Pod, blockingPod *BlockingPod, err error) {
	pods, daemonSetPods, blockingPod, err = GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		pdbs,
		false,
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		false,
		nil,
		0,
		time.Now())

	if err != nil {
		return pods, daemonSetPods, blockingPod, err
	}
	if pdbBlockingPod, err := checkPdbs(pods, pdbs); err != nil {
		return []*apiv1.Pod{}, []*apiv1.Pod{}, pdbBlockingPod, err
	}

	return pods, daemonSetPods, nil, nil
}

// DetailedGetPodsForMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod.
// Based on kubectl drain code. It checks whether RC, DS, Jobs and RS that created these pods
// still exist.
func DetailedGetPodsForMove(nodeInfo *schedulernodeinfo.NodeInfo, skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool, listers kube_util.ListerRegistry, minReplicaCount int32,
	pdbs []*policyv1.PodDisruptionBudget) (pods []*apiv1.Pod, daemonSetPods []*apiv1.Pod,
	blockingPod *BlockingPod, err error) {

	pods, daemonSetPods, blockingPod, err = GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		pdbs,
		false,
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		true,
		listers,
		minReplicaCount,
		time.Now())

	if err != nil {
		return pods, daemonSetPods, blockingPod, err
	}
	if pdbBlockingPod, err := checkPdbs(pods, pdbs); err != nil {
		return []*apiv1.Pod{}, []*apiv1.Pod{}, pdbBlockingPod, err
	}

	return pods, daemonSetPods, nil, nil
}

func checkPdbs(pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) (*BlockingPod, error) {
	// DOTO: make it more efficient.
	for _, pdb := range pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return nil, err
		}
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				if pdb.Status.PodDisruptionsAllowed < 1 {
					return &BlockingPod{Pod: pod, Reason: NotEnoughPdb},
						fmt.Errorf("not enough pod disruption budget to move %s/%s", pod.Namespace, pod.Name)
				}
			}
		}
	}
	return nil, nil
}

// HasLocalPV returns true if pod has any local pv.
func HasLocalPV(pod *apiv1.Pod, listers kube_util.ListerRegistry) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		if checkLocalPV(pod.Namespace, volume.VolumeSource, listers) {
			return true
		}
	}
	return false
}

// checkLocalPV xxx
// HasLocalPV returns true if pod has any local pv.
func checkLocalPV(ns string, vs apiv1.VolumeSource, listers kube_util.ListerRegistry) bool {
	if vs.PersistentVolumeClaim == nil {
		return false
	}
	pvcName := vs.PersistentVolumeClaim.ClaimName
	if listers == nil {
		return false
	}
	listersExtend, ok := listers.(util.ListerRegistryExtend)
	if !ok {
		return false
	}
	pvc, err := listersExtend.PVCLister().PersistentVolumeClaims(ns).Get(pvcName)
	if err != nil {
		return false
	}
	pv, err := listersExtend.PVLister().Get(pvc.Spec.VolumeName)
	if err != nil {
		return false
	}
	if pv.Spec.CSI == nil {
		return false
	}
	if pv.Spec.CSI.Driver == "localstorage.storage.csi.tencent.com" {
		return true
	}
	if pv.Spec.Local != nil {
		return true
	}
	return false
}

// IsPodLongTerminating checks if a pod has been terminating for a long time
// (pod's terminationGracePeriod + an additional const buffer)
func IsPodLongTerminating(pod *apiv1.Pod, currentTime time.Time) bool {
	// pod has not even been deleted
	if pod.DeletionTimestamp == nil {
		return false
	}

	gracePeriod := pod.Spec.TerminationGracePeriodSeconds
	if gracePeriod == nil {
		defaultGracePeriod := int64(apiv1.DefaultTerminationGracePeriodSeconds)
		gracePeriod = &defaultGracePeriod
	}
	return pod.DeletionTimestamp.Time.Add(time.Duration(*gracePeriod) * time.Second).Add(
		PodLongTerminatingExtraThreshold).Before(currentTime)
}

func checkRC(replicated *bool, checkReferences bool, pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	listers kube_util.ListerRegistry, minReplica int32) (*BlockingPod, error) {
	controllerNamespace := pod.Namespace
	if checkReferences {
		rc, err := listers.ReplicationControllerLister().ReplicationControllers(controllerNamespace).Get(controllerRef.Name)
		// Assume a reason for an error is because the RC is either
		// gone/missing or that the rc has too few replicas configured.
		// DOTO: replace the minReplica check with pod disruption budget.
		// nolint
		if err == nil && rc != nil {
			if rc.Spec.Replicas != nil && *rc.Spec.Replicas < minReplica {
				return &BlockingPod{Pod: pod, Reason: MinReplicasReached},
					fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d",
						pod.Namespace, pod.Name, rc.Spec.Replicas, minReplica)
			}
			*replicated = true
		} else {
			return &BlockingPod{Pod: pod, Reason: ControllerNotFound},
				fmt.Errorf("replication controller for %s/%s is not available, err: %v",
					pod.Namespace, pod.Name, err)
		}
	} else {
		*replicated = true
	}
	return nil, nil
}

// nolint `replicated` is unused
func checkDS(replicated *bool, checkReferences bool, pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	listers kube_util.ListerRegistry, minReplica int32) (*BlockingPod, error) {
	controllerNamespace := pod.Namespace
	if checkReferences {
		_, err := listers.DaemonSetLister().DaemonSets(controllerNamespace).Get(controllerRef.Name)
		// don't have listener for other DaemonSet kind
		if apierrors.IsNotFound(err) {
			return &BlockingPod{Pod: pod, Reason: ControllerNotFound},
				fmt.Errorf("daemonset for %s/%s is not present, err: %v", pod.Namespace, pod.Name, err)
		} else if err != nil {
			return &BlockingPod{Pod: pod, Reason: BlockingUnexpectedError},
				fmt.Errorf("error when trying to get daemonset for %s/%s , err: %v", pod.Namespace, pod.Name, err)
		}
	}
	return nil, nil
}

// nolint `minReplica` is unused
func checkJob(replicated *bool, checkReferences bool, pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	listers kube_util.ListerRegistry, minReplica int32) (*BlockingPod, error) {
	controllerNamespace := pod.Namespace
	if checkReferences {
		job, err := listers.JobLister().Jobs(controllerNamespace).Get(controllerRef.Name)

		// Assume the only reason for an error is because the Job is
		// gone/missing, not for any other cause.  DOTO(mml): something more
		// sophisticated than this
		// nolint
		if err == nil && job != nil {
			*replicated = true
		} else {
			return &BlockingPod{Pod: pod, Reason: ControllerNotFound},
				fmt.Errorf("job for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err)
		}
	} else {
		*replicated = true
	}
	return nil, nil
}

func checkRS(replicated *bool, checkReferences bool, pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	listers kube_util.ListerRegistry, minReplica int32) (*BlockingPod, error) {
	controllerNamespace := pod.Namespace
	if checkReferences {
		rs, err := listers.ReplicaSetLister().ReplicaSets(controllerNamespace).Get(controllerRef.Name)

		// Assume the only reason for an error is because the RS is
		// gone/missing, not for any other cause.  DOTO(mml): something more
		// sophisticated than this
		// nolint
		if err == nil && rs != nil {
			if rs.Spec.Replicas != nil && *rs.Spec.Replicas < minReplica {
				return &BlockingPod{Pod: pod, Reason: MinReplicasReached},
					fmt.Errorf("replication controller for %s/%s has too few replicas spec: %d min: %d",
						pod.Namespace, pod.Name, rs.Spec.Replicas, minReplica)
			}
			*replicated = true
		} else {
			return &BlockingPod{Pod: pod, Reason: ControllerNotFound},
				fmt.Errorf("replication controller for %s/%s is not available, err: %v", pod.Namespace, pod.Name, err)
		}
	} else {
		*replicated = true
	}
	return nil, nil
}

// nolint `minReplica` is unused
func checkSts(replicated *bool, checkReferences bool, pod *apiv1.Pod, controllerRef *metav1.OwnerReference,
	listers kube_util.ListerRegistry, minReplica int32) (*BlockingPod, error) {
	controllerNamespace := pod.Namespace
	if checkReferences {
		ss, err := listers.StatefulSetLister().StatefulSets(controllerNamespace).Get(controllerRef.Name)

		// Assume the only reason for an error is because the StatefulSet is
		// gone/missing, not for any other cause.  DOTO(mml): something more
		// sophisticated than this
		// nolint
		if err == nil && ss != nil {
			*replicated = true
		} else {
			return &BlockingPod{Pod: pod, Reason: ControllerNotFound},
				fmt.Errorf("statefulset for %s/%s is not available: err: %v", pod.Namespace, pod.Name, err)
		}
	} else {
		*replicated = true
	}
	return nil, nil
}
