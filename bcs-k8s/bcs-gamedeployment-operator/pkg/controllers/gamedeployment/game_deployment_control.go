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

package gamedeployment

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gamedeploylister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	revisioncontrol "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/revision"
	scalecontrol "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/scale"
	updatecontrol "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/update"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	canaryutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/canary"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/constants"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/controller/history"
)

// GameDeploymentControlInterface implements the control logic for updating GameDeployments and their children Pods. It is implemented
// as an interface to allow for extensions that provide different semantics. Currently, there is only one implementation.
type GameDeploymentControlInterface interface {
	// UpdateGameDeployment implements the control logic for Pod creation, update, and deletion.
	// If an implementation returns a non-nil error, the invocation will be retried using a rate-limited strategy.
	// Implementors should sink any errors that they do not wish to trigger a retry, and they may feel free to
	// exit exceptionally at any point provided they wish the update to be re-run at a later point in time.
	UpdateGameDeployment(deploy *v1alpha1.GameDeployment, pods []*v1.Pod) (time.Duration, *v1alpha1.GameDeploymentStatus, error)
	// ListRevisions returns a array of the ControllerRevisions that represent the revisions of set. If the returned
	// error is nil, the returns slice of ControllerRevisions is valid.
	ListRevisions(deploy *v1alpha1.GameDeployment) ([]*apps.ControllerRevision, error)
}

// NewDefaultGameDeploymentControl returns a new instance of the default implementation GameDeploymentControlInterface that
// implements the documented semantics for GameDeployments. statusUpdater is the GameDeploymentStatusUpdaterInterface used
// to update the status of GameDeployments.
func NewDefaultGameDeploymentControl(
	client tkexclientset.Interface,
	hookRunLister gamedeploylister.HookRunLister,
	hookTemplateLister gamedeploylister.HookTemplateLister,
	scaleControl scalecontrol.Interface,
	updateControl updatecontrol.Interface,
	statusUpdater GameDeploymentStatusUpdaterInterface,
	controllerHistory history.Interface,
	revisionControl revisioncontrol.Interface,
	recorder record.EventRecorder) GameDeploymentControlInterface {
	return &defaultGameDeploymentControl{
		client,
		scaleControl,
		updateControl,
		statusUpdater,
		controllerHistory,
		revisionControl,
		recorder,
		hookRunLister,
		hookTemplateLister,
	}
}

type defaultGameDeploymentControl struct {
	client             tkexclientset.Interface
	scaleControl       scalecontrol.Interface
	updateControl      updatecontrol.Interface
	statusUpdater      GameDeploymentStatusUpdaterInterface
	controllerHistory  history.Interface
	revisionControl    revisioncontrol.Interface
	recorder           record.EventRecorder
	hookRunLister      gamedeploylister.HookRunLister
	hookTemplateLister gamedeploylister.HookTemplateLister
}

func (gdc *defaultGameDeploymentControl) UpdateGameDeployment(deploy *v1alpha1.GameDeployment, pods []*v1.Pod) (time.Duration, *v1alpha1.GameDeploymentStatus, error) {
	if deploy.DeletionTimestamp != nil {
		return 0, nil, nil
	}

	key := fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("error converting GameDeployment %v selector: %v", key, err))
		// This is a non-transient error, so don't retry.
		return 0, nil, nil
	}

	// list all revisions and sort them
	revisions, err := gdc.controllerHistory.ListControllerRevisions(deploy, selector)
	if err != nil {
		return 0, nil, err
	}
	history.SortControllerRevisions(revisions)

	// get the current, and update revisions
	currentRevision, updateRevision, collisionCount, err := gdc.getActiveRevisions(deploy, revisions, util.GetPodsRevisions(pods))
	if err != nil {
		return 0, nil, err
	}

	// Refresh update expectations
	for _, pod := range pods {
		updateExpectations.ObserveUpdated(key, updateRevision.Name, pod)
	}
	// If update expectations have not satisfied yet, just skip this reconcile.
	if updateSatisfied, updateDirtyPods := updateExpectations.SatisfiedExpectations(key, updateRevision.Name); !updateSatisfied {
		klog.V(4).Infof("Not satisfied update for %v, updateDirtyPods=%v", key, updateDirtyPods)
		return 0, nil, nil
	}

	hrList, err := gdc.getHookRunsForGameDeployment(deploy)
	if err != nil {
		return 0, nil, err
	}

	canaryCtx := newCanaryCtx(deploy, hrList, updateRevision, collisionCount, selector)

	if canaryutil.CheckRevisionChange(deploy, updateRevision.Name) {
		err = gdc.statusUpdater.UpdateGameDeploymentStatus(deploy, canaryCtx, pods)
		return 0, canaryCtx.newStatus, err
	}
	if canaryutil.CheckStepHashChange(deploy) {
		err = gdc.statusUpdater.UpdateGameDeploymentStatus(deploy, canaryCtx, pods)
		return 0, canaryCtx.newStatus, err
	}

	err = gdc.reconcileHookRuns(canaryCtx)
	if err != nil {
		return 0, canaryCtx.newStatus, err
	}
	if canaryCtx.HasAddPause() {
		err = gdc.statusUpdater.UpdateGameDeploymentStatus(deploy, canaryCtx, pods)
		return 0, canaryCtx.newStatus, err
	}

	// scale and update pods
	delayDuration, updateErr := gdc.updateGameDeployment(deploy, canaryCtx.newStatus, currentRevision, updateRevision, revisions, pods)
	if updateErr != nil {
		return 0, canaryCtx.newStatus, err
	}

	unPauseDuration := gdc.reconcilePause(deploy)

	// update new status
	if err = gdc.statusUpdater.UpdateGameDeploymentStatus(deploy, canaryCtx, pods); err != nil {
		return 0, canaryCtx.newStatus, err
	}

	if err = gdc.truncatePodsToDelete(deploy, pods); err != nil {
		klog.Warningf("Failed to truncate podsToDelete for %s: %v", key, err)
	}

	if err = gdc.truncateHistory(deploy, pods, revisions, currentRevision, updateRevision); err != nil {
		klog.Errorf("Failed to truncate history for %s: %v", key, err)
	}

	// get a min duration between delayDuration and unPauseDuration
	requeueDuration := canaryutil.GetMinDuration(delayDuration, unPauseDuration)

	return requeueDuration, canaryCtx.newStatus, nil
}

func (gdc *defaultGameDeploymentControl) updateGameDeployment(
	deploy *v1alpha1.GameDeployment, newStatus *v1alpha1.GameDeploymentStatus,
	currentRevision, updateRevision *apps.ControllerRevision, revisions []*apps.ControllerRevision,
	filteredPods []*v1.Pod) (time.Duration, error) {

	var delayDuration time.Duration
	if deploy.DeletionTimestamp != nil {
		return delayDuration, nil
	}

	// get the current and update revisions of the set.
	currentSet, err := gdc.revisionControl.ApplyRevision(deploy, currentRevision)
	if err != nil {
		return delayDuration, err
	}
	updateSet, err := gdc.revisionControl.ApplyRevision(deploy, updateRevision)
	if err != nil {
		return delayDuration, err
	}

	var scaling bool
	var podsScaleErr error
	var podsUpdateErr error

	scaling, podsScaleErr = gdc.scaleControl.Manage(currentSet, updateSet, currentRevision.Name, updateRevision.Name, filteredPods)
	if podsScaleErr != nil {
		newStatus.Conditions = append(newStatus.Conditions, v1alpha1.GameDeploymentCondition{
			Type:               v1alpha1.GameDeploymentConditionFailedScale,
			Status:             v1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Message:            podsScaleErr.Error(),
		})
		err = podsScaleErr
	}
	if scaling {
		return delayDuration, podsScaleErr
	}

	delayDuration, podsUpdateErr = gdc.updateControl.Manage(updateSet, updateRevision, revisions, filteredPods)
	if podsUpdateErr != nil {
		newStatus.Conditions = append(newStatus.Conditions, v1alpha1.GameDeploymentCondition{
			Type:               v1alpha1.GameDeploymentConditionFailedUpdate,
			Status:             v1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Message:            podsUpdateErr.Error(),
		})
		if err == nil {
			err = podsUpdateErr
		}
	}

	return delayDuration, err
}

func (gdc *defaultGameDeploymentControl) reconcilePause(deploy *v1alpha1.GameDeployment) time.Duration {
	var timeRemaining time.Duration
	currentStep, _ := canaryutil.GetCurrentCanaryStep(deploy)
	if currentStep == nil || currentStep.Pause == nil || currentStep.Pause.Duration == nil {
		return timeRemaining
	}
	pauseCondition := canaryutil.GetPauseCondition(deploy, v1alpha1.PauseReasonCanaryPauseStep)
	if pauseCondition != nil {
		now := metav1.Now()
		expiredTime := pauseCondition.StartTime.Add(time.Duration(*currentStep.Pause.Duration) * time.Second)
		if expiredTime.After(now.Time) {
			timeRemaining = expiredTime.Sub(now.Time)
		}
	}
	return timeRemaining
}

func (gdc *defaultGameDeploymentControl) ListRevisions(deploy *v1alpha1.GameDeployment) ([]*apps.ControllerRevision, error) {
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return gdc.controllerHistory.ListControllerRevisions(deploy, selector)
}

func (gdc *defaultGameDeploymentControl) getActiveRevisions(deploy *v1alpha1.GameDeployment, revisions []*apps.ControllerRevision, podsRevisions sets.String) (
	*apps.ControllerRevision, *apps.ControllerRevision, int32, error,
) {
	var currentRevision, updateRevision *apps.ControllerRevision
	revisionCount := len(revisions)

	// Use a local copy of gd.Status.CollisionCount to avoid modifying gd.Status directly.
	// This copy is returned so the value gets carried over to gd.Status in UpdateGameDeploymentStatus.
	var collisionCount int32
	if deploy.Status.CollisionCount != nil {
		collisionCount = *deploy.Status.CollisionCount
	}

	// create a new revision from the current gd
	updateRevision, err := gdc.revisionControl.NewRevision(deploy, util.NextRevision(revisions), &collisionCount)
	if err != nil {
		return nil, nil, collisionCount, err
	}

	// find any equivalent revisions
	equalRevisions := history.FindEqualRevisions(revisions, updateRevision)
	equalCount := len(equalRevisions)

	if equalCount > 0 && history.EqualRevision(revisions[revisionCount-1], equalRevisions[equalCount-1]) {
		// if the equivalent revision is immediately prior the update revision has not changed
		updateRevision = revisions[revisionCount-1]
	} else if equalCount > 0 {
		// if the equivalent revision is not immediately prior we will roll back by incrementing the
		// Revision of the equivalent revision
		updateRevision, err = gdc.controllerHistory.UpdateControllerRevision(equalRevisions[equalCount-1], updateRevision.Revision)
		if err != nil {
			return nil, nil, collisionCount, err
		}
	} else {
		//if there is no equivalent revision we create a new one
		updateRevision, err = gdc.controllerHistory.CreateControllerRevision(deploy, updateRevision, &collisionCount)
		if err != nil {
			return nil, nil, collisionCount, err
		}
	}

	// attempt to find the revision that corresponds to the current revision
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

// truncatePodsToDelete truncates any non-live pod names in spec.scaleStrategy.podsToDelete.
func (gdc *defaultGameDeploymentControl) truncatePodsToDelete(deploy *v1alpha1.GameDeployment, pods []*v1.Pod) error {
	if len(deploy.Spec.ScaleStrategy.PodsToDelete) == 0 {
		return nil
	}

	existingPods := sets.NewString()
	for _, p := range pods {
		existingPods.Insert(p.Name)
	}

	var newPodsToDelete []string
	for _, podName := range deploy.Spec.ScaleStrategy.PodsToDelete {
		if existingPods.Has(podName) {
			newPodsToDelete = append(newPodsToDelete, podName)
		}
	}

	if len(newPodsToDelete) == len(deploy.Spec.ScaleStrategy.PodsToDelete) {
		return nil
	}

	newDeploy := deploy.DeepCopy()
	newDeploy.Spec.ScaleStrategy.PodsToDelete = newPodsToDelete
	_, updateErr := gdc.client.TkexV1alpha1().GameDeployments(deploy.Namespace).Update(newDeploy)
	return updateErr
}

// truncateHistory truncates any non-live ControllerRevisions in revisions from set's history. The UpdateRevision and
// CurrentRevision in set's Status are considered to be live. Any revisions associated with the Pods in pods are also
// considered to be live. Non-live revisions are deleted, starting with the revision with the lowest Revision, until
// only RevisionHistoryLimit revisions remain. If the returned error is nil the operation was successful. This method
// expects that revisions is sorted when supplied.
func (gdc *defaultGameDeploymentControl) truncateHistory(
	deploy *v1alpha1.GameDeployment,
	pods []*v1.Pod,
	revisions []*apps.ControllerRevision,
	current *apps.ControllerRevision,
	update *apps.ControllerRevision,
) error {
	history := make([]*apps.ControllerRevision, 0, len(revisions))
	// mark all live revisions
	live := map[string]bool{current.Name: true, update.Name: true}
	for i := range pods {
		live[util.GetPodRevision(pods[i])] = true
	}
	// collect live revisions and historic revisions
	for i := range revisions {
		if !live[revisions[i].Name] {
			history = append(history, revisions[i])
		}
	}
	historyLen := len(history)

	if deploy.Spec.RevisionHistoryLimit == nil {
		deploy.Spec.RevisionHistoryLimit = new(int32)
		*deploy.Spec.RevisionHistoryLimit = constants.DefaultRevisionHistoryLimit
	}

	historyLimit := int(*deploy.Spec.RevisionHistoryLimit)
	if historyLen <= historyLimit {
		return nil
	}
	// delete any non-live history to maintain the revision limit.
	history = history[:(historyLen - historyLimit)]
	for i := 0; i < len(history); i++ {
		if err := gdc.controllerHistory.DeleteControllerRevision(history[i]); err != nil {
			return err
		}
	}
	return nil
}
