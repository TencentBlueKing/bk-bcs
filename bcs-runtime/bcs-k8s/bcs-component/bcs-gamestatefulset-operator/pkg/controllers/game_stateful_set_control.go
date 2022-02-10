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
	"math"
	"sort"
	"time"

	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/util"
	canaryutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/util/canary"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/util/constants"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	hooklister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/postinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/preinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/hotpatchupdate"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/sets"
	clientset "k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller/history"
	"k8s.io/utils/integer"
)

// GameStatefulSetControlInterface implements the control logic for updating StatefulSets and their children Pods. It is implemented
// as an interface to allow for extensions that provide different semantics. Currently, there is only one implementation.
type GameStatefulSetControlInterface interface {
	// UpdateGameStatefulSet implements the control logic for Pod creation, update, and deletion, and
	// persistent volume creation, update, and deletion.
	// If an implementation returns a non-nil error, the invocation will be retried using a rate-limited strategy.
	// Implementors should sink any errors that they do not wish to trigger a retry, and they may feel free to
	// exit exceptionally at any point provided they wish the update to be re-run at a later point in time.
	UpdateGameStatefulSet(set *gstsv1alpha1.GameStatefulSet, pods []*v1.Pod) error
	// ListRevisions returns a array of the ControllerRevisions that represent the revisions of set. If the returned
	// error is nil, the returns slice of ControllerRevisions is valid.
	ListRevisions(set *gstsv1alpha1.GameStatefulSet) ([]*apps.ControllerRevision, error)
	// AdoptOrphanRevisions adopts any orphaned ControllerRevisions that match set's Selector. If all adoptions are
	// successful the returned error is nil.
	AdoptOrphanRevisions(set *gstsv1alpha1.GameStatefulSet, revisions []*apps.ControllerRevision) error
}

// NewDefaultGameStatefulSetControl returns a new instance of the default implementation StatefulSetControlInterface that
// implements the documented semantics for StatefulSets. podControl is the PodControlInterface used to create, update,
// and delete Pods and to create PersistentVolumeClaims. statusUpdater is the StatefulSetStatusUpdaterInterface used
// to update the status of StatefulSets. You should use an instance returned from NewRealStatefulPodControl() for any
// scenario other than testing.
func NewDefaultGameStatefulSetControl(
	kubeClient clientset.Interface,
	hookClient hookclientset.Interface,
	podControl GameStatefulSetPodControlInterface,
	inPlaceControl inplaceupdate.Interface,
	hotPatchControl hotpatchupdate.Interface,
	statusUpdater GameStatefulSetStatusUpdaterInterface,
	controllerHistory history.Interface,
	recorder record.EventRecorder,
	podLister corelisters.PodLister,
	hookRunLister hooklister.HookRunLister,
	hookTemplateLister hooklister.HookTemplateLister,
	preDeleteControl predelete.PreDeleteInterface,
	preInplaceControl preinplace.PreInplaceInterface,
	postInplaceControl postinplace.PostInplaceInterface,
	metrics *metrics) GameStatefulSetControlInterface {
	return &defaultGameStatefulSetControl{
		kubeClient,
		hookClient,
		podControl,
		statusUpdater,
		controllerHistory,
		recorder,
		inPlaceControl,
		hotPatchControl,
		podLister,
		hookRunLister,
		hookTemplateLister,
		preDeleteControl,
		preInplaceControl,
		postInplaceControl,
		metrics,
	}
}

type defaultGameStatefulSetControl struct {
	kubeClient         clientset.Interface
	hookClient         hookclientset.Interface
	podControl         GameStatefulSetPodControlInterface
	statusUpdater      GameStatefulSetStatusUpdaterInterface
	controllerHistory  history.Interface
	recorder           record.EventRecorder
	inPlaceControl     inplaceupdate.Interface
	hotPatchControl    hotpatchupdate.Interface
	podLister          corelisters.PodLister
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
	preDeleteControl   predelete.PreDeleteInterface
	preInplaceControl  preinplace.PreInplaceInterface
	postInplaceControl postinplace.PostInplaceInterface
	// metrics is used to collect prom metrics
	metrics *metrics
}

// UpdateGameStatefulSet executes the core logic loop for a stateful set plus, applying the predictable and
// consistent monotonic update strategy by default - scale up proceeds in ordinal order, no new pod
// is created while any pod is unhealthy, and pods are terminated in descending order. The burst
// strategy allows these constraints to be relaxed - pods will be created and deleted eagerly and
// in no particular order. Clients using the burst strategy should be careful to ensure they
// understand the consistency implications of having unpredictable numbers of pods available.
func (ssc *defaultGameStatefulSetControl) UpdateGameStatefulSet(
	set *gstsv1alpha1.GameStatefulSet,
	pods []*v1.Pod) error {

	// list all revisions and sort them
	revisions, err := ssc.ListRevisions(set)
	if err != nil {
		klog.Errorf("List GameStatefulSet %s/%s relative ControllerRevision err: %#v", set.Namespace, set.Name, err)
		return err
	}
	history.SortControllerRevisions(revisions)
	// get the current, and update revisions
	currentRevision, updateRevision, collisionCount, err := ssc.getGameStatefulSetRevisions(
		set, revisions, getPodsRevisions(pods))
	if err != nil {
		return err
	}

	hrList, err := ssc.getHookRunsForGameStatefulSet(set)
	if err != nil {
		return err
	}

	canaryCtx := newCanaryCtx(set, hrList, currentRevision, updateRevision, collisionCount)

	if canaryutil.CheckRevisionChange(set, updateRevision.Name) {
		err = ssc.updateGameStatefulSetStatus(set, canaryCtx)
		return err
	}
	if canaryutil.CheckStepHashChange(set) {
		err = ssc.updateGameStatefulSetStatus(set, canaryCtx)
		return err
	}

	err = ssc.reconcileHookRuns(canaryCtx)
	if err != nil {
		return err
	}
	if canaryCtx.HasAddPause() {
		err = ssc.updateGameStatefulSetStatus(set, canaryCtx)
		return err
	}

	// perform the main update function and get the status
	_, updateErr := ssc.updateGameStatefulSet(
		set,
		canaryCtx.newStatus,
		currentRevision,
		updateRevision,
		pods,
		revisions,
		hrList)
	if updateErr != nil {
		return err
	}

	// delete scale down dirty pods whose hooks are completed
	key := fmt.Sprintf("%s/%s", set.Namespace, set.Name)
	scaleDirtyPods := scaleExpectations.GetExpectations(key)
	ssc.handleDirtyPods(set, canaryCtx.newStatus, scaleDirtyPods[expectations.Delete].List())

	unPauseDuration := ssc.reconcilePause(set)
	if unPauseDuration > 0 {
		durationStore.Push(getGameStatefulSetKey(set), unPauseDuration)
	}

	// update the set's status
	err = ssc.updateGameStatefulSetStatus(set, canaryCtx)
	if err != nil {
		return err
	}
	ssc.metrics.collectRelatedReplicas(set.Namespace, set.Name, set.Status.Replicas, set.Status.ReadyReplicas,
		set.Status.CurrentReplicas, set.Status.UpdatedReplicas, set.Status.UpdatedReadyReplicas)

	// klog.V(3).Infof("GameStatefulSet %s/%s pod status replicas=%d ready=%d current=%d updated=%d",
	// 	set.Namespace,
	// 	set.Name,
	// 	status.Replicas,
	// 	status.ReadyReplicas,
	// 	status.CurrentReplicas,
	// 	status.UpdatedReplicas)

	// klog.V(3).Infof("GameStatefulSet %s/%s revisions current=%s update=%s",
	// 	set.Namespace,
	// 	set.Name,
	// 	status.CurrentRevision,
	// 	status.UpdateRevision)

	// maintain the set's revision history limit
	return ssc.truncateHistory(set, pods, revisions, currentRevision, updateRevision)
}

func (ssc *defaultGameStatefulSetControl) reconcilePause(set *gstsv1alpha1.GameStatefulSet) time.Duration {
	var timeRemaining time.Duration
	currentStep, _ := canaryutil.GetCurrentCanaryStep(set)
	if currentStep == nil || currentStep.Pause == nil || currentStep.Pause.Duration == nil {
		return timeRemaining
	}
	pauseCondition := canaryutil.GetPauseCondition(set, hookv1alpha1.PauseReasonCanaryPauseStep)
	if pauseCondition != nil {
		now := metav1.Now()
		expiredTime := pauseCondition.StartTime.Add(time.Duration(*currentStep.Pause.Duration) * time.Second)
		if expiredTime.After(now.Time) {
			timeRemaining = expiredTime.Sub(now.Time)
		}
	}
	return timeRemaining
}

// ListRevisions list all revisions of gamestatefulset
func (ssc *defaultGameStatefulSetControl) ListRevisions(
	set *gstsv1alpha1.GameStatefulSet) ([]*apps.ControllerRevision, error) {
	selector, err := metav1.LabelSelectorAsSelector(set.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return ssc.controllerHistory.ListControllerRevisions(set, selector)
}

// AdoptOrphanRevisions adopt orphan history revision of gamestatefulset
func (ssc *defaultGameStatefulSetControl) AdoptOrphanRevisions(
	set *gstsv1alpha1.GameStatefulSet,
	revisions []*apps.ControllerRevision) error {
	for i := range revisions {
		adopted, err := ssc.controllerHistory.AdoptControllerRevision(set, util.ControllerKind, revisions[i])
		if err != nil {
			return err
		}
		revisions[i] = adopted
	}
	return nil
}

// truncateHistory truncates any non-live ControllerRevisions in revisions from set's history. The UpdateRevision and
// CurrentRevision in set's Status are considered to be live. Any revisions associated with the Pods in pods are also
// considered to be live. Non-live revisions are deleted, starting with the revision with the lowest Revision, until
// only RevisionHistoryLimit revisions remain. If the returned error is nil the operation was successful. This method
// expects that revisions is sorted when supplied.
func (ssc *defaultGameStatefulSetControl) truncateHistory(
	set *gstsv1alpha1.GameStatefulSet,
	pods []*v1.Pod,
	revisions []*apps.ControllerRevision,
	current *apps.ControllerRevision,
	update *apps.ControllerRevision) error {
	history := make([]*apps.ControllerRevision, 0, len(revisions))
	// mark all live revisions
	live := map[string]bool{current.Name: true, update.Name: true}
	for i := range pods {
		live[getPodRevision(pods[i])] = true
	}
	// collect live revisions and historic revisions
	for i := range revisions {
		if !live[revisions[i].Name] {
			history = append(history, revisions[i])
		}
	}
	historyLen := len(history)

	if set.Spec.RevisionHistoryLimit == nil {
		set.Spec.RevisionHistoryLimit = new(int32)
		*set.Spec.RevisionHistoryLimit = constants.DefaultRevisionHistoryLimit
	}

	historyLimit := int(*set.Spec.RevisionHistoryLimit)
	if historyLen <= historyLimit {
		return nil
	}
	// delete any non-live history to maintain the revision limit.
	history = history[:(historyLen - historyLimit)]
	for i := 0; i < len(history); i++ {
		if err := ssc.controllerHistory.DeleteControllerRevision(history[i]); err != nil {
			return err
		}
	}
	return nil
}

// getStatefulSetRevisions returns the current and update ControllerRevisions for set. It also
// returns a collision count that records the number of name collisions set saw when creating
// new ControllerRevisions. This count is incremented on every name collision and is used in
// building the ControllerRevision names for name collision avoidance. This method may create
// a new revision, or modify the Revision of an existing revision if an update to set is detected.
// This method expects that revisions is sorted when supplied.
func (ssc *defaultGameStatefulSetControl) getGameStatefulSetRevisions(
	set *gstsv1alpha1.GameStatefulSet,
	revisions []*apps.ControllerRevision,
	podsRevisions sets.String) (*apps.ControllerRevision, *apps.ControllerRevision, int32, error) {
	var currentRevision, updateRevision *apps.ControllerRevision
	revisionCount := len(revisions)
	history.SortControllerRevisions(revisions)

	klog.V(4).Infof("getGameStatefulSetRevisions for %s/%s, revision number: %d",
		set.GetNamespace(), set.GetName(), revisionCount)

	// Use a local copy of set.Status.CollisionCount to avoid modifying set.Status directly.
	// This copy is returned so the value gets carried over to set.Status in updateGameStatefulSet.
	var collisionCount int32
	if set.Status.CollisionCount != nil {
		collisionCount = *set.Status.CollisionCount
	}

	// create a new revision from the current set
	updateRevision, err := newRevision(set, nextRevision(revisions), &collisionCount)
	if err != nil {
		klog.Errorf("create new revision for %s/%s err: %#v, collision count: %d",
			set.GetNamespace(), set.GetName(), err, collisionCount)
		return nil, nil, collisionCount, err
	}

	// find any equivalent revisions
	equalRevisions := history.FindEqualRevisions(revisions, updateRevision)
	equalCount := len(equalRevisions)
	if equalCount > 0 && history.EqualRevision(revisions[revisionCount-1], equalRevisions[equalCount-1]) {
		// if the equivalent revision is immediately prior the update revision has not changed
		updateRevision = revisions[revisionCount-1]
		klog.V(4).Infof("GameStatefulSet %s/%s revision is same with prior revision, nothing change.",
			set.GetNamespace(), set.GetName())
	} else if equalCount > 0 {
		// if the equivalent revision is not immediately prior we will roll back by incrementing the
		// Revision of the equivalent revision
		updateRevision, err = ssc.controllerHistory.UpdateControllerRevision(
			equalRevisions[equalCount-1],
			updateRevision.Revision)
		if err != nil {
			klog.Errorf("update equal controllerRevision %s/%s to %d err: %#v",
				set.GetNamespace(), set.GetName(), updateRevision.Revision, err)
			return nil, nil, collisionCount, err
		}
		klog.V(4).Infof("update previous controllerRevision %s/%s from %d to %d successfully",
			set.GetNamespace(), set.GetName(), equalRevisions[equalCount-1].Revision, updateRevision.Revision)
	} else {
		//if there is no equivalent revision we create a new one
		updateRevision, err = ssc.controllerHistory.CreateControllerRevision(set, updateRevision, &collisionCount)
		if err != nil {
			klog.Errorf("create new controllerRevision for %s/%s err: %#v", set.GetNamespace(), set.GetName(), err)
			return nil, nil, collisionCount, err
		}
		klog.V(3).Infof("create new controllerRevision %d/%s for %s/%s successfully",
			updateRevision.Revision, updateRevision.Name, set.GetNamespace(), set.GetName())
	}

	for i := range revisions {
		if podsRevisions.Has(revisions[i].Name) {
			currentRevision = revisions[i]
			break
		}
	}

	// if the current revision is nil we initialize the history by setting it to the update revision
	if currentRevision == nil {
		currentRevision = updateRevision
	}

	return currentRevision, updateRevision, collisionCount, nil
}

// updateGameStatefulSet performs the update function for a GameStatefulSet. This method creates, updates, and deletes Pods in
// the set in order to conform the system to the target state for the set. The target state always contains
// set.Spec.Replicas Pods with a Ready Condition. If the UpdateStrategy.Type for the set is
// RollingUpdateGameStatefulSetStrategyType then all Pods in the set must be at set.Status.CurrentRevision.
// If the UpdateStrategy.Type for the set is OnDeleteGameStatefulSetStrategyType, the target state implies nothing about
// the revisions of Pods in the set. If the UpdateStrategy.Type for the set is PartitionGameStatefulSetStrategyType, then
// all Pods with ordinal less than UpdateStrategy.Partition.Ordinal must be at Status.CurrentRevision and all other
// Pods must be at Status.UpdateRevision. If the returned error is nil, the returned StatefulSetStatus is valid and the
// update must be recorded. If the error is not nil, the method should be retried until successful.
func (ssc *defaultGameStatefulSetControl) updateGameStatefulSet(
	set *gstsv1alpha1.GameStatefulSet,
	status *gstsv1alpha1.GameStatefulSetStatus,
	currentRevision *apps.ControllerRevision,
	updateRevision *apps.ControllerRevision,
	pods []*v1.Pod,
	revisions []*apps.ControllerRevision,
	hrList []*hookv1alpha1.HookRun) (*gstsv1alpha1.GameStatefulSetStatus, error) {
	// get the current and update revisions of the set.
	currentSet, err := ApplyRevision(set, currentRevision)
	if err != nil {
		return nil, err
	}
	updateSet, err := ApplyRevision(set, updateRevision)
	if err != nil {
		return nil, err
	}

	// truncate unneeded PreDeleteHookRuns
	err = ssc.truncatePreDeleteHookRuns(set, pods, hrList)
	if err != nil {
		return status, err
	}

	ssc.truncatePreDeleteHookConditions(pods, status)

	// if configured retry, delete unexpected HookRuns and reconcile
	if set.Spec.PreDeleteUpdateStrategy.RetryUnexpectedHooks {
		return status, ssc.deleteUnexpectedPreDeleteHookRuns(hrList)
	}

	// truncate unneeded PreDeleteHookRuns
	err = ssc.truncatePreInplaceHookRuns(set, pods, hrList)
	if err != nil {
		return status, err
	}

	ssc.truncatePreInplaceHookConditions(pods, status)

	// if configured retry, delete unexpected HookRuns and reconcile
	if set.Spec.PreInplaceUpdateStrategy.RetryUnexpectedHooks {
		return status, ssc.deleteUnexpectedPreInplaceHookRuns(hrList)
	}

	replicaCount := int(*set.Spec.Replicas)
	// slice that will contain all Pods such that 0 <= getOrdinal(pod) < set.Spec.Replicas
	replicas := make([]*v1.Pod, replicaCount)
	// slice that will contain all Pods such that set.Spec.Replicas <= getOrdinal(pod)
	condemned := make([]*v1.Pod, 0, len(pods))
	unhealthy := 0
	firstUnhealthyOrdinal := math.MaxInt32
	var firstUnhealthyPod *v1.Pod

	updatedReadyOfReplicasCount := 0

	// First we partition pods into two lists valid replicas and condemned Pods
	for i := range pods {
		status.Replicas++

		// count the number of running and ready replicas
		if isRunningAndReady(pods[i]) {
			status.ReadyReplicas++
		}

		if isRunningAndReady(pods[i]) && getPodRevision(pods[i]) == updateRevision.Name {
			status.UpdatedReadyReplicas++
		}

		// count the number of running, ready, and updated pods of replicasCount
		if ord := getOrdinal(pods[i]); 0 <= ord && ord < replicaCount && isRunningAndReady(pods[i]) &&
			getPodRevision(pods[i]) == updateRevision.Name {
			updatedReadyOfReplicasCount++
		}

		// count the number of current and update replicas
		if isCreated(pods[i]) && !isTerminating(pods[i]) {
			ssc.renewStatus(status, pods[i], currentRevision, updateRevision, 1)
		}

		if ord := getOrdinal(pods[i]); 0 <= ord && ord < replicaCount {
			// if the ordinal of the pod is within the range of the current number of replicas,
			// insert it at the indirection of its ordinal
			replicas[ord] = pods[i]

		} else if ord >= replicaCount {
			// if the ordinal is greater than the number of replicas add it to the condemned list
			condemned = append(condemned, pods[i])
		}
		// If the ordinal could not be parsed (ord < 0), ignore the Pod.
	}

	// sort the condemned Pods by their ordinals
	sort.Sort(ascendingOrdinal(condemned))

	replicas, condemned, err = ssc.dealWithMaxSurge(set, currentRevision, updateRevision,
		replicaCount, updatedReadyOfReplicasCount, replicas, condemned)
	if err != nil {
		return status, err
	}

	// for any empty indices in the sequence [0,set.Spec.Replicas) create a new Pod at the correct revision
	for ord := 0; ord < len(replicas); ord++ {
		if replicas[ord] == nil {
			replicas[ord] = newVersionedGameStatefulSetPod(
				set,
				currentSet,
				updateSet,
				currentRevision.Name,
				updateRevision.Name, ord)
		}
	}

	// find the first unhealthy Pod
	for i := range replicas {
		if !isHealthy(replicas[i]) {
			unhealthy++
			if ord := getOrdinal(replicas[i]); ord < firstUnhealthyOrdinal {
				firstUnhealthyOrdinal = ord
				firstUnhealthyPod = replicas[i]
			}
		}
	}

	for i := range condemned {
		if !isHealthy(condemned[i]) {
			unhealthy++
			if ord := getOrdinal(condemned[i]); ord < firstUnhealthyOrdinal {
				firstUnhealthyOrdinal = ord
				firstUnhealthyPod = condemned[i]
			}
		}
	}

	if unhealthy > 0 {
		klog.Infof("GameStatefulSet %s/%s has %d unhealthy Pods starting with %s",
			set.Namespace,
			set.Name,
			unhealthy,
			firstUnhealthyPod.Name)
	}

	// If the GameStatefulSet is being deleted, don't do anything other than updating
	// status.
	if set.DeletionTimestamp != nil {
		return status, nil
	}

	// resync post inplace hook
	ssc.updatePostInplaceHookConditions(replicas, set, status)

	monotonic := !allowsBurst(set)

	// Examine each replica with respect to its ordinal
	for i := range replicas {
		// delete and recreate failed pods
		if isFailed(replicas[i]) {
			ssc.recorder.Eventf(set, v1.EventTypeWarning, "RecreatingFailedPod",
				"GameStatefulSet %s/%s is recreating failed Pod %s",
				set.Namespace,
				set.Name,
				replicas[i].Name)
			klog.Infof("GameStatefulSet %s/%s is deleting failed Pod %s and then recreating",
				set.Namespace, set.Name, replicas[i].Name)
			if err := ssc.podControl.DeleteGameStatefulSetPod(set, replicas[i]); err != nil {
				klog.Errorf("Operator delete Pod %s controlled by GameStatefulSet %s/%s failed, %s",
					replicas[i].Name, set.Namespace, set.Name, err.Error())
				return status, err
			}
			ssc.renewStatus(status, replicas[i], currentRevision, updateRevision, -1)
			status.Replicas--
			replicas[i] = newVersionedGameStatefulSetPod(
				set,
				currentSet,
				updateSet,
				currentRevision.Name,
				updateRevision.Name,
				i)
		}

		//new feature: force delete and recreate NodeLost pods
		if isTerminating(replicas[i]) {
			deleted, err := ssc.podControl.ForceDeleteGameStatefulSetPod(set, replicas[i])
			if err != nil {
				klog.Errorf("Operator force delete Pod %s controlled by GameStatefulSet %s/%s failed, %s",
					replicas[i].Name, set.Namespace, set.Name, err.Error())
				return status, err
			}
			if deleted {
				ssc.renewStatus(status, replicas[i], currentRevision, updateRevision, -1)
				status.Replicas--
				replicas[i] = newVersionedGameStatefulSetPod(
					set,
					currentSet,
					updateSet,
					currentRevision.Name,
					updateRevision.Name,
					i)
			}
		}

		// If we find a Pod that has not been created we create the Pod
		if !isCreated(replicas[i]) {
			if err := ssc.podControl.CreateGameStatefulSetPod(set, replicas[i]); err != nil {
				klog.Errorf("Operator create new Pod %s controlled by GameStatefulSet %s/%s failed, %s",
					replicas[i].Name, set.Namespace, set.Name, err.Error())
				return status, err
			}
			klog.Infof("GameStatefulSet %s/%s is creating Pod %s", set.Namespace, set.Name, replicas[i].Name)
			status.Replicas++
			ssc.renewStatus(status, replicas[i], currentRevision, updateRevision, 1)

			// if the set does not allow bursting, return immediately
			if monotonic {
				return status, nil
			}
			// pod created, no more work possible for this round
			continue
		}
		// If we find a Pod that is currently terminating, we must wait until graceful deletion
		// completes before we continue to make progress.
		if isTerminating(replicas[i]) && monotonic {
			klog.V(3).Infof(
				"GameStatefulSet %s/%s is waiting for Pod %s to Terminate",
				set.Namespace,
				set.Name,
				replicas[i].Name)
			return status, nil
		}

		// If we find a Pod that is current terminating and PodManagmentPolicy is Parallel,
		// we should ignore it and continue check next replica.
		if isTerminating(replicas[i]) && !monotonic {
			klog.V(3).Infof(
				"GameStatefulSet Pod %s/%s/%s is terminating, and the PodManagmentPolicy is Parallel, "+
					"we will not wait until graceful deletition completes before we continue to make progress.",
				set.Namespace,
				set.Name,
				replicas[i].Name)
			continue
		}

		// Update InPlaceUpdateReady condition for pod
		if res := ssc.inPlaceControl.Refresh(replicas[i], nil); res.RefreshErr != nil {
			klog.Errorf("StatefulSet %s/%s failed to update pod %s condition for inplace: %v",
				set.Namespace, set.Name, replicas[i].Name, res.RefreshErr)
			return status, res.RefreshErr
		} else if res.DelayDuration > 0 {
			durationStore.Push(getGameStatefulSetKey(set), res.DelayDuration)
		}

		// If we have a Pod that has been created but is not running and ready we can not make progress.
		// We must ensure that all for each Pod, when we create it, all of its predecessors, with respect to its
		// ordinal, are Running and Ready.
		if monotonic && (getPodRevision(replicas[i]) == updateRevision.Name) && (!isRunningAndReady(replicas[i])) {
			klog.V(3).Infof(
				"GameStatefulSet %s/%s is waiting for Pod %s to be Running and Ready",
				set.Namespace,
				set.Name,
				replicas[i].Name)
			return status, nil
		}

		// Enforce the GameStatefulSet invariants
		if IdentityMatches(set, replicas[i]) && storageMatches(set, replicas[i]) {
			continue
		}
		// Make a deep copy so we don't mutate the shared cache
		replica := replicas[i].DeepCopy()
		startTime := time.Now()
		if err := ssc.podControl.UpdateGameStatefulSetPod(updateSet, replica); err != nil {
			klog.Errorf("Update GameStatefulSet %s/%s in normal replicas iteration err, %s",
				updateSet.Namespace, updateSet.Name, err.Error())
			ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "failure", time.Since(startTime))
			return status, err
		}
		ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "success", time.Since(startTime))
	}

	// At this point, all of the current Replicas are Running and Ready, we can consider termination.
	// We will wait for all predecessors to be Running and Ready prior to attempting a deletion.
	// We will terminate Pods in a monotonically decreasing order over [len(pods),set.Spec.Replicas).
	// Note that we do not resurrect Pods in this interval. Also not that scaling will take precedence over
	// updates.
	for target := len(condemned) - 1; target >= 0; target-- {
		// wait for terminating pods to expire
		if isTerminating(condemned[target]) {
			klog.V(3).Infof(
				"GameStatefulSet %s/%s is waiting for Pod %s to Terminate prior to scale down",
				set.Namespace,
				set.Name,
				condemned[target].Name)
			// block if we are in monotonic mode
			if monotonic {
				return status, nil
			}
			continue
		}
		// if we are in monotonic mode and the condemned target is not the first unhealthy Pod block
		if !isRunningAndReady(condemned[target]) && monotonic && condemned[target] != firstUnhealthyPod {
			klog.V(3).Infof(
				"GameStatefulSet %s/%s is waiting for Pod %s to be Running and Ready prior to scale down",
				set.Namespace,
				set.Name,
				firstUnhealthyPod.Name)
			return status, nil
		}

		scaleExpectations.ExpectScale(util.GetControllerKey(set), expectations.Delete, condemned[target].Name)
		canDelete, err := ssc.preDeleteControl.CheckDelete(
			set,
			condemned[target],
			status,
			gstsv1alpha1.GameStatefulSetPodOrdinal)
		if err != nil {
			klog.Errorf(
				"Error to check whether the pod %s can be safely deleted for GameStatefulSet %s/%s: %s",
				condemned[target].Name, set.Namespace, set.Name, err.Error())
			return status, err
		}
		if !canDelete {
			klog.V(2).Infof(
				"PreDelete Hook not completed, can't delete the pod %s for GameStatefulSet %s/%s now.",
				condemned[target].Name,
				set.Namespace,
				set.Name)
			if monotonic {
				return status, nil
			}
		} else {
			klog.Infof("GameStatefulSet %s/%s is terminating Pod %s for scale down",
				set.Namespace,
				set.Name,
				condemned[target].Name)

			if err := ssc.podControl.DeleteGameStatefulSetPod(set, condemned[target]); err != nil {
				klog.Errorf("GameStatefulSet %s/%s clean condemoned Pod %d err: %s",
					set.Namespace, set.Name, target, err.Error())
				return status, err
			}
			ssc.renewStatus(status, condemned[target], currentRevision, updateRevision, -1)
			if monotonic {
				return status, nil
			}
		}
	}

	return ssc.handleUpdateStrategy(set, status, revisions, updateRevision, replicas, monotonic)
}

func (ssc *defaultGameStatefulSetControl) handleUpdateStrategy(
	set *gstsv1alpha1.GameStatefulSet,
	status *gstsv1alpha1.GameStatefulSetStatus,
	revisions []*apps.ControllerRevision,
	updateRevision *apps.ControllerRevision,
	replicas []*v1.Pod,
	monotonic bool,
) (*gstsv1alpha1.GameStatefulSetStatus, error) {

	// for the OnDelete strategy we short circuit. Pods will be updated when they are manually deleted.
	if set.Spec.UpdateStrategy.Type == gstsv1alpha1.OnDeleteGameStatefulSetStrategyType {
		return status, nil
	}

	// we compute the minimum ordinal of the target sequence for a destructive update based on the strategy.
	updateMin := 0
	currentPartition, err := canaryutil.GetCurrentPartition(set)
	if err != nil {
		return status, nil
	}
	updateMin = int(currentPartition)

	replicasCount := int(*set.Spec.Replicas)
	maxUnavailable, err := intstrutil.GetValueFromIntOrPercent(intstrutil.ValueOrDefault(
		set.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable, intstrutil.FromString("25%")), replicasCount, true)
	if err != nil {
		return status, err
	}

	// Collect all targets in the range between the partition and Spec.Replicas.
	// Count any targets in that range that are unhealthy i.e. terminated or not running and ready as unavailable.
	// Select the (MaxUnavailable - Unavailable) Pods, in order with respect to their ordinal for termination.
	// Delete those pods and count the successful deletions. Update the status with the correct number of deletions.
	unavailablePods := 0
	for target := len(replicas) - 1; target >= updateMin; target-- {
		if !isHealthy(replicas[target]) {
			unavailablePods++
		}
	}
	// Now we need to delete (MaxUnavailable - unavailablePods) Pods,
	// start updating one by one starting from the highest ordinal first
	podsToUpdate := maxUnavailable - unavailablePods
	if podsToUpdate < 1 {
		// if unavailablePods >= MaxUnavailable, should not update any pods.
		return status, nil
	}
	updatedPods := 0

	switch set.Spec.UpdateStrategy.Type {
	case gstsv1alpha1.InplaceUpdateGameStatefulSetStrategyType:
		for target := replicasCount - 1; target >= updateMin; target-- {
			if getPodRevision(replicas[target]) != updateRevision.Name && !isTerminating(replicas[target]) {
				updatedPods++
				if set.Spec.PreInplaceUpdateStrategy.Hook != nil {
					canInplace, err := ssc.preInplaceControl.CheckInplace(
						set,
						replicas[target],
						&set.Spec.Template,
						status,
						gstsv1alpha1.GameStatefulSetPodOrdinal)
					if err != nil {
						klog.Errorf(
							"Error to check whether the pod %s can be safely deleted for gsts %s/%s: %s",
							replicas[target].Name,
							set.Namespace,
							set.Name,
							err.Error())
						return status, err
					}
					if !canInplace {
						klog.V(2).Infof(
							"PreInplace Hook not completed, can't inplaceUpdate the pod %s for gsts %s/%s now.",
							replicas[target].Name,
							set.Namespace,
							set.Name)

						if !ssc.continueUpdate(set, monotonic, maxUnavailable, unavailablePods, updatedPods, podsToUpdate) {
							return status, nil
						}
						continue
					}
				} else {
					canDelete, err := ssc.preDeleteControl.CheckDelete(
						set,
						replicas[target],
						status,
						gstsv1alpha1.GameStatefulSetPodOrdinal)
					if err != nil {
						klog.Errorf(
							"Error to check whether the pod %s can be safely deleted for gsts %s/%s: %s",
							replicas[target].Name, set.Namespace, set.Name, err.Error())
						return status, err
					}
					if !canDelete {
						klog.V(2).Infof(
							"PreDelete Hook not completed, can't inplaceUpdate the pod %s for gsts %s/%s now.",
							replicas[target].Name,
							set.Namespace,
							set.Name)
						if !ssc.continueUpdate(set, monotonic, maxUnavailable, unavailablePods, updatedPods, podsToUpdate) {
							return status, nil
						}
						continue
					}
				}
				inPlaceUpdateErr := ssc.inPlaceUpdatePod(set, replicas[target], updateRevision, revisions)
				status.CurrentReplicas--
				if inPlaceUpdateErr != nil {
					return status, inPlaceUpdateErr
				}

				// create post inplace hook
				newPod, err := ssc.kubeClient.CoreV1().Pods(replicas[target].Namespace).Get(context.TODO(), replicas[target].Name,
					metav1.GetOptions{})
				if err != nil {
					klog.Warningf("Cannot get pod %s/%s", replicas[target].Namespace, replicas[target].Name)
				} else {
					created, err := ssc.postInplaceControl.CreatePostInplaceHook(set,
						newPod,
						status,
						gstsv1alpha1.GameStatefulSetPodOrdinal)
					if err != nil {
						ssc.recorder.Eventf(set, v1.EventTypeWarning, "FailedCreatePostHookRun",
							"failed to create post inplace hook for pod %s, error: %v", replicas[target].Name, err)
					} else if created {
						ssc.recorder.Eventf(set, v1.EventTypeNormal, "SuccessfulCreatePostHookRun",
							"successfully create post inplace hook for pod %s", replicas[target].Name)
					} else {
						ssc.recorder.Eventf(set, v1.EventTypeNormal, "PostHookRunExisted",
							"post inplace hook for pod %s has been existed", replicas[target].Name)
					}
				}

				if !ssc.continueUpdate(set, monotonic, maxUnavailable, unavailablePods, updatedPods, podsToUpdate) {
					return status, nil
				}
				continue
			}

			if !isHealthy(replicas[target]) && monotonic {
				klog.V(3).Infof(
					"GameStatefulSet %s/%s is waiting for Pod %s healthy to InplaceUpdate",
					set.Namespace,
					set.Name,
					replicas[target].Name)
				return status, nil
			}
		}
	// RollingUpdate handle here
	case gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType:
		// we terminate the Pod with the largest ordinal that does not match the update revision.
		for target := replicasCount - 1; target >= updateMin; target-- {

			// delete the Pod if it is not already terminating and does not match the update revision.
			if getPodRevision(replicas[target]) != updateRevision.Name && !isTerminating(replicas[target]) {
				updatedPods++
				canDelete, err := ssc.preDeleteControl.CheckDelete(
					set,
					replicas[target],
					status,
					gstsv1alpha1.GameStatefulSetPodOrdinal)
				if err != nil {
					klog.Errorf(
						"Error to check whether the pod %s can be safely deleted for GameStatefulSet %s/%s: %s",
						replicas[target].Name, set.Namespace, set.Name, err.Error())
					return status, err
				}
				if !canDelete {
					klog.V(2).Infof(
						"PreDelete Hook not completed, can't delete the pod %s for GameStatefulSet %s/%s now.",
						replicas[target].Name,
						set.Namespace,
						set.Name)
					if !ssc.continueUpdate(set, monotonic, maxUnavailable, unavailablePods, updatedPods, podsToUpdate) {
						return status, nil
					}
					continue
				} else {
					klog.Infof("GameStatefulSet %s/%s terminating Pod %s for RollingUpdate",
						set.Namespace,
						set.Name,
						replicas[target].Name)
					err := ssc.podControl.DeleteGameStatefulSetPod(set, replicas[target])
					status.CurrentReplicas--
					if err != nil {
						return status, err
					}
					if !ssc.continueUpdate(set, monotonic, maxUnavailable, unavailablePods, updatedPods, podsToUpdate) {
						return status, nil
					}
					continue
				}
			}

			// wait for unhealthy Pods on update
			if !isHealthy(replicas[target]) && monotonic {
				klog.V(3).Infof(
					"GameStatefulSet %s/%s is waiting for Pod %s to RollingUpdate",
					set.Namespace,
					set.Name,
					replicas[target].Name)
				return status, nil
			}
		}

	//(DeveloperBryan): InplaceHotPatch handle here
	case gstsv1alpha1.HotPatchGameStatefulSetStrategyType:
		for target := replicasCount - 1; target >= updateMin; target-- {
			if getPodRevision(replicas[target]) != updateRevision.Name {
				klog.Infof("GameStatefulSet %s/%s updating Pod %s to InplaceUpdate", set.Namespace,
					set.Name, replicas[target].Name)

				err := ssc.hotPatchUpdatePod(set, replicas[target], updateRevision, revisions)
				status.CurrentReplicas--
				return status, err
			}

			if !isHealthy(replicas[target]) && monotonic {
				klog.V(3).Infof(
					"GameStatefulSet %s/%s is waiting for Pod %s healthy to InplaceUpdate",
					set.Namespace,
					set.Name,
					replicas[target].Name)
				return status, nil
			}
		}
	}

	return status, nil
}

func (ssc *defaultGameStatefulSetControl) inPlaceUpdatePod(
	set *gstsv1alpha1.GameStatefulSet, pod *v1.Pod,
	updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
) error {
	var oldRevision *apps.ControllerRevision
	for _, r := range revisions {
		if r.Name == getPodRevision(pod) {
			oldRevision = r
			break
		}
	}

	opts := &inplaceupdate.UpdateOptions{}
	if set.Spec.UpdateStrategy.InPlaceUpdateStrategy != nil {
		opts.GracePeriodSeconds = set.Spec.UpdateStrategy.InPlaceUpdateStrategy.GracePeriodSeconds
	}

	startTime := time.Now()
	res := ssc.inPlaceControl.Update(pod, oldRevision, updateRevision, opts)

	if res.UpdateErr == nil {
		ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "success", time.Since(startTime))
	} else {
		ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "failure", time.Since(startTime))
	}

	if res.InPlaceUpdate {
		if res.UpdateErr == nil {
			ssc.recorder.Eventf(
				set,
				v1.EventTypeNormal,
				"SuccessfulUpdatePodInPlace",
				"successfully update pod %s in-place",
				pod.Name)
		} else {
			ssc.recorder.Eventf(
				set,
				v1.EventTypeWarning,
				"FailedUpdatePodInPlace",
				"failed to update pod %s in-place: %v",
				pod.Name,
				res.UpdateErr)
		}
		if res.DelayDuration > 0 {
			durationStore.Push(getGameStatefulSetKey(set), res.DelayDuration)
		}
		return res.UpdateErr
	}

	err := fmt.Errorf(
		"find Pod %s update strategy is InplaceUpdate, but the diff "+
			"not only contains replace operation of spec.containers[x].image",
		pod)
	ssc.recorder.Eventf(
		set,
		v1.EventTypeWarning,
		"FailedUpdatePodInPlace",
		"find Pod %s update strategy is InPlace but can not update in-place: %v",
		pod.Name,
		err)
	klog.Warningf("GameStatefulSet %s/%s can not update Pod %s in-place: %v", set.Namespace, set.Name, pod.Name, err)
	return err
}

func (ssc *defaultGameStatefulSetControl) hotPatchUpdatePod(
	set *gstsv1alpha1.GameStatefulSet, pod *v1.Pod,
	updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
) error {
	var oldRevision *apps.ControllerRevision
	for _, r := range revisions {
		if r.Name == getPodRevision(pod) {
			oldRevision = r
			break
		}
	}

	startTime := time.Now()
	err := ssc.hotPatchControl.Update(pod, oldRevision, updateRevision)

	if err == nil {
		ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "success", time.Since(startTime))
	} else {
		ssc.metrics.collectPodUpdateDurations(set.Namespace, set.Name, "failure", time.Since(startTime))
	}

	if err != nil {
		ssc.recorder.Eventf(
			set,
			v1.EventTypeWarning,
			"FailedUpdatePodHotPatch",
			"failed to update pod %s hot-patch: %v",
			pod.Name,
			err)
		return err
	}
	ssc.recorder.Eventf(
		set,
		v1.EventTypeNormal,
		"SuccessfulUpdatePodHotPatch",
		"successfully update pod %s hot-patch",
		pod.Name)
	return nil
}

// updateGameStatefulSetStatus updates set's Status to be equal to status. If status indicates a complete update, it is
// mutated to indicate completion. If status is semantically equivalent to set's Status no update is performed. If the
// returned error is nil, the update is successful.
func (ssc *defaultGameStatefulSetControl) updateGameStatefulSetStatus(
	set *gstsv1alpha1.GameStatefulSet,
	canaryCtx *canaryContext) error {

	// copy set and update its status
	set = set.DeepCopy()
	if err := ssc.statusUpdater.UpdateGameStatefulSetStatus(set, canaryCtx); err != nil {
		return err
	}

	return nil
}

// truncatePreDeleteHookRuns truncate unneeded PreDeleteHookConditions
func (ssc *defaultGameStatefulSetControl) truncatePreDeleteHookRuns(
	set *gstsv1alpha1.GameStatefulSet,
	pods []*v1.Pod,
	hrList []*hookv1alpha1.HookRun) error {
	preDeleteHookRuns := commonhookutil.FilterPreDeleteHookRuns(hrList)
	hrsToDelete := []*hookv1alpha1.HookRun{}
	for _, hr := range preDeleteHookRuns {
		podControllerRevision := hr.Labels[commonhookutil.WorkloadRevisionUniqueLabel]
		podInstanceID := hr.Labels[commonhookutil.PodInstanceID]

		exist := false
		for _, pod := range pods {
			cr, ok1 := pod.Labels[apps.ControllerRevisionHashLabelKey]
			id, ok2 := pod.Labels[gstsv1alpha1.GameStatefulSetPodOrdinal]
			if ok1 && ok2 && podControllerRevision == cr && podInstanceID == id {
				exist = true
				break
			}
		}
		if !exist {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return ssc.deleteHookRuns(hrsToDelete)
}

// truncatePreDeleteHookConditions truncate unneeded PreDeleteHookConditions
func (ssc *defaultGameStatefulSetControl) truncatePreDeleteHookConditions(
	pods []*v1.Pod,
	newStatus *gstsv1alpha1.GameStatefulSetStatus) {
	tmpPredeleteHookConditions := []hookv1alpha1.PreDeleteHookCondition{}
	for _, cond := range newStatus.PreDeleteHookConditions {
		for _, pod := range pods {
			if cond.PodName == pod.Name {
				tmpPredeleteHookConditions = append(tmpPredeleteHookConditions, cond)
				break
			}
		}
		newStatus.PreDeleteHookConditions = tmpPredeleteHookConditions
	}
}

// deleteUnexpectedPreDeleteHookRuns delete unexpected PreDeleteHookRuns, then will trigger a reconcile
func (ssc *defaultGameStatefulSetControl) deleteUnexpectedPreDeleteHookRuns(hrList []*hookv1alpha1.HookRun) error {
	preDeleteHookRuns := commonhookutil.FilterPreDeleteHookRuns(hrList)
	hrsToDelete := []*hookv1alpha1.HookRun{}
	for _, hr := range preDeleteHookRuns {
		if hr.Status.Phase.Completed() && hr.Status.Phase != hookv1alpha1.HookPhaseSuccessful {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return ssc.deleteHookRuns(hrsToDelete)
}

// truncatePreInplaceHookRuns truncate unneeded PreInplaceHookConditions
func (ssc *defaultGameStatefulSetControl) truncatePreInplaceHookRuns(
	set *gstsv1alpha1.GameStatefulSet,
	pods []*v1.Pod,
	hrList []*hookv1alpha1.HookRun) error {
	preInplaceHookRuns := commonhookutil.FilterPreInplaceHookRuns(hrList)
	hrsToDelete := []*hookv1alpha1.HookRun{}
	for _, hr := range preInplaceHookRuns {
		podControllerRevision := hr.Labels[commonhookutil.WorkloadRevisionUniqueLabel]
		podInstanceID := hr.Labels[commonhookutil.PodInstanceID]

		exist := false
		for _, pod := range pods {
			cr, ok1 := pod.Labels[apps.ControllerRevisionHashLabelKey]
			id, ok2 := pod.Labels[gstsv1alpha1.GameStatefulSetPodOrdinal]
			if ok1 && ok2 && podControllerRevision == cr && podInstanceID == id {
				exist = true
				break
			}
		}
		if !exist {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return ssc.deleteHookRuns(hrsToDelete)
}

// truncatePreInplaceHookConditions truncate unneeded PreInplaceHookConditions
func (ssc *defaultGameStatefulSetControl) truncatePreInplaceHookConditions(
	pods []*v1.Pod,
	newStatus *gstsv1alpha1.GameStatefulSetStatus) {
	tmpPreInplaceHookConditions := []hookv1alpha1.PreInplaceHookCondition{}
	for _, cond := range newStatus.PreInplaceHookConditions {
		for _, pod := range pods {
			if cond.PodName == pod.Name {
				tmpPreInplaceHookConditions = append(tmpPreInplaceHookConditions, cond)
				break
			}
		}
		newStatus.PreInplaceHookConditions = tmpPreInplaceHookConditions
	}
}

// deleteUnexpectedPreInplaceHookRuns delete unexpected PreInplaceHookRuns, then will trigger a reconcile
func (ssc *defaultGameStatefulSetControl) deleteUnexpectedPreInplaceHookRuns(
	hrList []*hookv1alpha1.HookRun) error {
	preInplaceHookRuns := commonhookutil.FilterPreInplaceHookRuns(hrList)
	hrsToDelete := []*hookv1alpha1.HookRun{}
	for _, hr := range preInplaceHookRuns {
		if hr.Status.Phase.Completed() && hr.Status.Phase != hookv1alpha1.HookPhaseSuccessful {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return ssc.deleteHookRuns(hrsToDelete)
}

var _ GameStatefulSetControlInterface = &defaultGameStatefulSetControl{}

func (ssc *defaultGameStatefulSetControl) updatePostInplaceHookConditions(
	replicas []*v1.Pod,
	set *gstsv1alpha1.GameStatefulSet,
	status *gstsv1alpha1.GameStatefulSetStatus) {
	for _, pod := range replicas {
		err := ssc.postInplaceControl.UpdatePostInplaceHook(set, pod, status, gstsv1alpha1.GameStatefulSetPodOrdinal)
		if err != nil {
			ssc.recorder.Eventf(set, v1.EventTypeWarning, "FailedResyncPostHookRun",
				"failed to resync post hook for pod %s, error: %v", pod.Name, err)
		}
	}
}

func (ssc *defaultGameStatefulSetControl) handleDirtyPods(set *gstsv1alpha1.GameStatefulSet,
	newStatus *gstsv1alpha1.GameStatefulSetStatus, dirtyPods []string) {
	for _, podName := range dirtyPods {
		err := ssc.deletePod(set, newStatus, podName)
		if err != nil {
			klog.Infof("Failed to delete pod %s/%s: %s", set.Namespace, podName, err.Error())
		}
	}
}

func (ssc *defaultGameStatefulSetControl) deletePod(set *gstsv1alpha1.GameStatefulSet,
	newStatus *gstsv1alpha1.GameStatefulSetStatus, podName string) error {
	pod, err := ssc.podLister.Pods(set.Namespace).Get(podName)
	if err != nil && errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		ssc.recorder.Eventf(set, v1.EventTypeWarning, "FailedGetPod",
			"failed to get pod %s/%s: %v", pod.Namespace, pod.Name, err)
		return err
	}

	canDelete, err := ssc.preDeleteControl.CheckDelete(set, pod, newStatus, gstsv1alpha1.GameStatefulSetPodOrdinal)
	if err != nil {
		klog.V(2).Infof("CheckDelete failed for pod %s/%s: %v", pod.Namespace, pod.Name, err)
		return err
	}
	if canDelete {
		if set.Spec.PreDeleteUpdateStrategy.Hook != nil {
			klog.V(2).Infof("PreDelete Hook run successfully, delete the pod %s/%s now.", pod.Namespace, pod.Name)
		}
	} else {
		klog.V(2).Infof("PreDelete Hook not completed, can't delete the pod %s/%s now.", pod.Namespace, pod.Name)
		return fmt.Errorf("PreDelete Hook of pod %s/%s not completed", pod.Namespace, pod.Name)
	}
	startTime := time.Now()
	if err := ssc.kubeClient.CoreV1().Pods(pod.Namespace).Delete(context.TODO(),
		pod.Name, metav1.DeleteOptions{}); err != nil {
		scaleExpectations.ObserveScale(util.GetControllerKey(set), expectations.Delete, pod.Name)
		ssc.recorder.Eventf(set, v1.EventTypeWarning, "FailedDeletePod",
			"failed to delete pod %s/%s: %v", set.Namespace, podName, err)
		ssc.metrics.collectPodDeleteDurations(pod.Namespace, pod.Name, "failure", time.Since(startTime))
		return err
	}
	ssc.metrics.collectPodDeleteDurations(pod.Namespace, pod.Name, "success", time.Since(startTime))
	return nil
}

// continueUpdate determines whether to update next pod
func (ssc *defaultGameStatefulSetControl) continueUpdate(set *gstsv1alpha1.GameStatefulSet,
	monotonic bool, maxUnavailable, unavailablePods, updatedPods, podsToUpdate int) bool {
	// if monotonic, stop updating more pods
	if monotonic {
		return false
	}

	// If at anytime, total number of unavailable Pods exceeds maxUnavailable,
	// we stop updating more Pods
	if updatedPods == podsToUpdate {
		klog.V(4).Infof(
			"GameStatefulSet is waiting for pods to become available: gst=%s/%s, maxUnavailable=%d, "+
				"unavailablePods=%d, operatedPods=%d",
			set.Namespace,
			set.Name,
			maxUnavailable,
			unavailablePods,
			updatedPods,
		)
		return false
	}

	return true
}

// dealWithMaxSurge returns replicas and condemned after taking maxSurge into accounts.
func (ssc *defaultGameStatefulSetControl) dealWithMaxSurge(set *gstsv1alpha1.GameStatefulSet,
	currentRevision, updateRevision *apps.ControllerRevision,
	replicaCount, updatedReadyOfReplicasCount int,
	replicas, condemned []*v1.Pod) ([]*v1.Pod, []*v1.Pod, error) {
	// Only use maxSurge when updating pods in parallel and strategy is not OnDelete.
	if !allowsBurst(set) {
		return replicas, condemned, nil
	}
	if !isNotOnDeleteUpdate(set) {
		return replicas, condemned, nil
	}
	if currentRevision == updateRevision {
		return replicas, condemned, nil
	}

	// When considerating maxSurge, we should tempporarily increase replicaCount
	// to (relicaCount + min(maxSurge, notUpdatedCounts)), as well as change replicas and condemned.
	maxSurge, err := intstrutil.GetValueFromIntOrPercent(intstrutil.ValueOrDefault(
		set.Spec.UpdateStrategy.RollingUpdate.MaxSurge, intstrutil.FromInt(0)), replicaCount, true)
	if err != nil {
		return replicas, condemned, err
	}
	partition, err := intstrutil.GetValueFromIntOrPercent(set.Spec.UpdateStrategy.RollingUpdate.Partition, replicaCount, true)
	if err != nil {
		return replicas, condemned, err
	}
	// reserve (partition) currentRevision pods
	notUpdatedCounts := replicaCount - updatedReadyOfReplicasCount - partition
	maxSurge = integer.IntMin(maxSurge, notUpdatedCounts)
	// when partition >= replicasCounts, do nothing
	if maxSurge > 0 && maxSurge <= len(condemned) {
		replicas = append(replicas, condemned[:maxSurge]...)
		condemned = condemned[maxSurge:]
	} else if maxSurge > 0 {
		replicas = append(replicas, condemned...)
		for i := 0; i < maxSurge-len(condemned); i++ {
			replicas = append(replicas, nil)
		}
		condemned = []*v1.Pod{}
	}

	return replicas, condemned, nil
}

func (ssc *defaultGameStatefulSetControl) renewStatus(status *gstsv1alpha1.GameStatefulSetStatus,
	pod *v1.Pod, currentRevision, updateRevision *apps.ControllerRevision, num int) {
	if getPodRevision(pod) == currentRevision.Name {
		status.CurrentReplicas = status.CurrentReplicas + int32(num)
	}
	if getPodRevision(pod) == updateRevision.Name {
		status.UpdatedReplicas = status.UpdatedReplicas + int32(num)
	}
}
