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

package gamestatefulset

import (
	"context"
	"errors"
	stsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	stsFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/clientset/versioned/fake"
	stsscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/clientset/versioned/scheme"
	stsInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	v1alpha12 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/postinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/preinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/hotpatchupdate"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestGetGameStatefulSetRevisions(t *testing.T) {
	_ = stsscheme.AddToScheme(scheme.Scheme)
	var collisionCount int32

	// initialize test data
	sts1 := testutil.NewGameStatefulSet(1)
	// because revision will hash the spec.template, so we need to change the spec.template
	sts1.Spec.Template.Labels["test"] = "test1"
	dRev1, err := newRevision(sts1, 1, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	sts2 := testutil.NewGameStatefulSet(2)
	sts2.Spec.Template.Labels["test"] = "test2"
	dRev2, err := newRevision(sts2, 2, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	sts3 := testutil.NewGameStatefulSet(3)
	sts3.Spec.Template.Labels["test"] = "test3"
	dRev3, err := newRevision(sts3, 3, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	dRev4, err := newRevision(sts2, 4, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		sts          *stsv1alpha1.GameStatefulSet
		revisions    []*apps.ControllerRevision
		podRevisions sets.String

		expectedCurrentRevision *apps.ControllerRevision
		expectedUpdateRevision  *apps.ControllerRevision
		expectedCollisionCount  int32
		expectedError           error
	}{
		{ // the equivalent revision is the latest revision
			name:                    "the equivalent revision is the latest revision",
			sts:                     sts3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev2.Name: {}},
			expectedCurrentRevision: dRev2,
			expectedUpdateRevision:  dRev3,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // the equivalent revision isn't the latest revision
			name:                    "the equivalent revision isn't the latest revision",
			sts:                     sts2,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev3.Name: {}},
			expectedCurrentRevision: dRev3,
			expectedUpdateRevision:  dRev4,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // haven't equivalent revision
			name:                    "haven't equivalent revision",
			sts:                     sts3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2},
			podRevisions:            map[string]sets.Empty{dRev1.Name: {}},
			expectedCurrentRevision: dRev1,
			expectedUpdateRevision:  dRev3,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // when initializing, the latest revision is the current revision
			name:                    "when initializing",
			sts:                     sts3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2},
			podRevisions:            map[string]sets.Empty{},
			expectedCurrentRevision: dRev3,
			expectedUpdateRevision:  dRev3,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			stop := make(chan struct{})
			defer close(stop)
			informerFactory.Start(stop)
			informer := informerFactory.Apps().V1().ControllerRevisions()
			informerFactory.WaitForCacheSync(stop)
			for i := range s.revisions {
				informer.Informer().GetIndexer().Add(s.revisions[i])
			}
			controllerHistory := history.NewFakeHistory(informer)
			control := &defaultGameStatefulSetControl{controllerHistory: controllerHistory}

			currentRevision, updateRevision, collisionCount, err := control.getGameStatefulSetRevisions(s.sts, s.revisions, s.podRevisions)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !reflect.DeepEqual(currentRevision, s.expectedCurrentRevision) {
				t.Errorf("expected current revision %v, got %v", s.expectedCurrentRevision, currentRevision)
			}
			if !reflect.DeepEqual(updateRevision, s.expectedUpdateRevision) {
				t.Errorf("expected update revision %v, got %v", s.expectedUpdateRevision, updateRevision)
			}
			if collisionCount != s.expectedCollisionCount {
				t.Errorf("expected collision count %v, got %v", s.expectedCollisionCount, collisionCount)
			}
		})
	}
}

func TestDeleteUnexpectedPreDeleteHookRuns(t *testing.T) {
	tests := []struct {
		name            string
		hrList          []*v1alpha12.HookRun
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "one delete",
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
		{
			name: "two hr, one was deleted",
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
				newHR("hr2", v1alpha12.HookPhaseFailed, true, ""),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hookClient := hookFake.NewSimpleClientset()
			gdc := &defaultGameStatefulSetControl{
				hookClient: hookClient,
			}
			err := gdc.deleteUnexpectedPreDeleteHookRuns(s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !testutil.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestTruncatePreDeleteHookRuns(t *testing.T) {
	tests := []struct {
		name            string
		pods            []*corev1.Pod
		hrList          []*v1alpha12.HookRun
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "not exist",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
		{
			name: "exist",
			pods: []*corev1.Pod{
				func() *corev1.Pod {
					pod := testutil.NewPod(0)
					pod.Labels[apps.ControllerRevisionHashLabelKey] = "1"
					pod.Labels[stsv1alpha1.GameStatefulSetPodOrdinal] = "1"
					return pod
				}(),
				testutil.NewPod(1),
			},
			hrList: []*v1alpha12.HookRun{
				func() *v1alpha12.HookRun {
					hr := newHR("hr1", v1alpha12.HookPhaseFailed, false, "")
					hr.Labels[commonhookutil.WorkloadRevisionUniqueLabel] = "1"
					hr.Labels[commonhookutil.PodInstanceID] = "1"
					return hr
				}(),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hookClient := hookFake.NewSimpleClientset()
			gdc := &defaultGameStatefulSetControl{
				hookClient: hookClient,
			}
			err := gdc.truncatePreDeleteHookRuns(testutil.NewGameStatefulSet(1), s.pods, s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !testutil.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestDeleteUnexpectedPreInplaceHookRuns(t *testing.T) {
	tests := []struct {
		name            string
		hrList          []*v1alpha12.HookRun
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "one delete",
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, commonhookutil.HookRunTypePreInplaceLabel),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
		{
			name: "two hr, one was deleted",
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, commonhookutil.HookRunTypePreInplaceLabel),
				newHR("hr2", v1alpha12.HookPhaseFailed, true, commonhookutil.HookRunTypePreInplaceLabel),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hookClient := hookFake.NewSimpleClientset()
			gdc := &defaultGameStatefulSetControl{
				hookClient: hookClient,
			}
			err := gdc.deleteUnexpectedPreInplaceHookRuns(s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !testutil.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestTruncatePreInplaceHookRuns(t *testing.T) {
	tests := []struct {
		name            string
		pods            []*corev1.Pod
		hrList          []*v1alpha12.HookRun
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "not exist",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			hrList: []*v1alpha12.HookRun{
				newHR("hr1", v1alpha12.HookPhaseFailed, false, commonhookutil.HookRunTypePreInplaceLabel),
			},
			expectedActions: []testing2.Action{
				testing2.NewDeleteAction(runtimeSchema.GroupVersionResource{
					Group:    "tkex.tencent.com",
					Version:  "v1alpha1",
					Resource: "hookruns",
				}, "default", "hr1"),
			},
		},
		{
			name: "exist",
			pods: []*corev1.Pod{
				func() *corev1.Pod {
					pod := testutil.NewPod(0)
					pod.Labels[apps.ControllerRevisionHashLabelKey] = "1"
					pod.Labels[stsv1alpha1.GameStatefulSetPodOrdinal] = "1"
					return pod
				}(),
				testutil.NewPod(1),
			},
			hrList: []*v1alpha12.HookRun{
				func() *v1alpha12.HookRun {
					hr := newHR("hr1", v1alpha12.HookPhaseFailed, false, commonhookutil.HookRunTypePreInplaceLabel)
					hr.Labels[commonhookutil.WorkloadRevisionUniqueLabel] = "1"
					hr.Labels[commonhookutil.PodInstanceID] = "1"
					return hr
				}(),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hookClient := hookFake.NewSimpleClientset()
			gdc := &defaultGameStatefulSetControl{
				hookClient: hookClient,
			}
			err := gdc.truncatePreInplaceHookRuns(testutil.NewGameStatefulSet(1), s.pods, s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !testutil.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestGDCDeletePod(t *testing.T) {
	// Create the controller
	kubeClient := fake.NewSimpleClientset()
	hookClient := hookFake.NewSimpleClientset()
	kubeInformer := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
	hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	recorder := record.NewFakeRecorder(1000)

	gdc := &defaultGameStatefulSetControl{
		kubeClient: kubeClient,
		recorder:   recorder,
		podLister:  kubeInformer.Core().V1().Pods().Lister(),
		preDeleteControl: predelete.New(kubeClient, hookClient, recorder,
			hookInformer.Tkex().V1alpha1().HookRuns().Lister(), hookInformer.Tkex().V1alpha1().HookTemplates().Lister()),
		metrics: newMetrics(),
	}

	pod := testutil.NewPod(0)
	_ = kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
	gdc.deletePod(testutil.NewGameStatefulSet(1), &stsv1alpha1.GameStatefulSetStatus{}, pod.Name)
	if got, want := len(kubeClient.Actions()), 1; got != want {
		t.Fatalf("not expected pod actions count, want: %d, got: %d", want, got)
	}
	if !kubeClient.Actions()[0].Matches("delete", "pods") {
		t.Errorf("not expected pod actions verb")
	}
	kubeClient.ClearActions()

	// test pod not exist
	pod2 := testutil.NewPod(2)
	gdc.deletePod(testutil.NewGameStatefulSet(1), &stsv1alpha1.GameStatefulSetStatus{}, pod2.Name)
	if got, want := len(kubeClient.Actions()), 0; got != want {
		t.Fatalf("not expected pod actions count, want: %d, got: %d", want, got)
	}
}

func newSts(limit int32) *stsv1alpha1.GameStatefulSet {
	d := testutil.NewGameStatefulSet(1)
	d.Spec.RevisionHistoryLimit = &limit
	return d
}

func newPodWithControllerRevision(revision string) *corev1.Pod {
	pod := testutil.NewPod(1)
	pod.Labels[apps.ControllerRevisionHashLabelKey] = revision
	return pod
}

func newControllerRevision(name string) *apps.ControllerRevision {
	return &apps.ControllerRevision{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func TestTruncateHistory(t *testing.T) {
	tests := []struct {
		name                 string
		revisionHistoryLimit int32
		pods                 []*corev1.Pod
		revisions            []*apps.ControllerRevision
		currentRevisionName  string
		updateRevisionName   string
		expectedError        error
		expectedRemainKeys   []string
	}{
		{
			name:                 "normal case",
			revisionHistoryLimit: 1,
			pods: []*corev1.Pod{
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("1"),
				newControllerRevision("2"),
				newControllerRevision("3"),
			},
			currentRevisionName: "2",
			updateRevisionName:  "4",
			expectedRemainKeys:  []string{"2", "3"},
		},
		{
			name:                 "limit more than history count",
			revisionHistoryLimit: 3,
			pods: []*corev1.Pod{
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("1"),
				newControllerRevision("2"),
				newControllerRevision("3"),
			},
			currentRevisionName: "2",
			updateRevisionName:  "4",
			expectedRemainKeys:  []string{"1", "2", "3"},
		},
		{
			name:                 "unsort revisions",
			revisionHistoryLimit: 1,
			pods: []*corev1.Pod{
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("3"),
				newControllerRevision("1"),
				newControllerRevision("2"),
			},
			currentRevisionName: "2",
			updateRevisionName:  "4",
			expectedRemainKeys:  []string{"1", "2"},
		},
		{
			name:                 "revision history limit is 0",
			revisionHistoryLimit: 0,
			pods: []*corev1.Pod{
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("2"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("3"),
				newControllerRevision("1"),
				newControllerRevision("2"),
			},
			currentRevisionName: "2",
			updateRevisionName:  "4",
			expectedRemainKeys:  []string{"2"},
		},
		{
			name:                 "more revision",
			revisionHistoryLimit: 2,
			pods: []*corev1.Pod{
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("1"),
				newPodWithControllerRevision("2"),
				newPodWithControllerRevision("3"),
				newPodWithControllerRevision("2"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("1"),
				newControllerRevision("2"),
				newControllerRevision("3"),
				newControllerRevision("4"),
				newControllerRevision("5"),
				newControllerRevision("6"),
				newControllerRevision("7"),
			},
			currentRevisionName: "6",
			updateRevisionName:  "8",
			expectedRemainKeys:  []string{"1", "2", "3", "5", "6", "7"},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()
			kubeInformer := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
			gdc := &defaultGameStatefulSetControl{
				controllerHistory: history.NewFakeHistory(kubeInformer.Apps().V1().ControllerRevisions()),
			}
			for _, controllerRevision := range s.revisions {
				kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().Add(controllerRevision)
			}
			err := gdc.truncateHistory(newSts(s.revisionHistoryLimit), s.pods, s.revisions,
				newControllerRevision(s.currentRevisionName), newControllerRevision(s.updateRevisionName))
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			keys := kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().ListKeys()
			sort.Strings(keys)
			sort.Strings(s.expectedRemainKeys)
			if !reflect.DeepEqual(keys, s.expectedRemainKeys) {
				t.Errorf("expected remain keys %v, got %v", s.expectedRemainKeys, keys)
			}
		})
	}
}

func TestReconcilePause(t *testing.T) {
	tests := []struct {
		name                  string
		set                   *stsv1alpha1.GameStatefulSet
		expectedTimeRemaining time.Duration
	}{
		{
			name: "no pause step",
			set:  testutil.NewGameStatefulSet(1),
		},
		{
			name: "expired after now",
			set: func() *stsv1alpha1.GameStatefulSet {
				set := testutil.NewGameStatefulSet(1)
				set.Spec.UpdateStrategy.CanaryStrategy = &stsv1alpha1.CanaryStrategy{Steps: []stsv1alpha1.CanaryStep{
					{Pause: &stsv1alpha1.CanaryPause{Duration: func() *int32 { i := int32(1000000000); return &i }()}},
				}}
				set.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.UnixMilli(1600000000000)),
					},
				}
				return set
			}(),
			expectedTimeRemaining: 1,
		},
		{
			name: "expired before now",
			set: func() *stsv1alpha1.GameStatefulSet {
				set := testutil.NewGameStatefulSet(1)
				set.Spec.UpdateStrategy.CanaryStrategy = &stsv1alpha1.CanaryStrategy{Steps: []stsv1alpha1.CanaryStep{
					{Pause: &stsv1alpha1.CanaryPause{Duration: func() *int32 { i := int32(1); return &i }()}},
				}}
				set.Status.PauseConditions = []v1alpha12.PauseCondition{
					{
						Reason:    v1alpha12.PauseReasonCanaryPauseStep,
						StartTime: metav1.NewTime(time.UnixMilli(1600000000000)),
					},
				}
				return set
			}(),
			expectedTimeRemaining: 0,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			ssc := &defaultGameStatefulSetControl{}
			timeRemaining := ssc.reconcilePause(s.set)
			// cause the time now to be different, so we can test the time remaining with '=='
			if s.expectedTimeRemaining > timeRemaining {
				t.Errorf("expected time remaining %v, got %v", s.expectedTimeRemaining, timeRemaining)
			}
		})
	}
}

func newPodWithStatus(ordinal int, phase corev1.PodPhase, revision string, ready, delete bool) *corev1.Pod {
	pod := testutil.NewPod(ordinal)
	pod.Status.Phase = phase
	pod.Labels[stsv1alpha1.GameStatefulSetRevisionLabel] = revision
	pod.Spec.Hostname = pod.Name
	pod.Spec.ReadinessGates = []corev1.PodReadinessGate{
		{ConditionType: "InPlaceUpdateReady"},
	}
	if ready {
		pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
			Type:   corev1.PodReady,
			Status: corev1.ConditionTrue,
		})
		pod.Status.Conditions = append(pod.Status.Conditions, corev1.PodCondition{
			Type:   "InPlaceUpdateReady",
			Status: corev1.ConditionTrue,
		})
	}
	if delete {
		pod.DeletionTimestamp = &metav1.Time{Time: time.Now()}
		pod.Annotations[podNodeLostForceDeleteKey] = "true"
		pod.Spec.NodeName = "foo"
	}
	return pod
}

func newCRWithName(name string, revision int64, sts *stsv1alpha1.GameStatefulSet) *apps.ControllerRevision {
	cr, err := newRevision(sts, revision, func() *int32 { i := int32(0); return &i }())
	if err != nil {
		return nil
	}
	cr.Name = name
	cr.Namespace = sts.Namespace
	cr.Labels["foo"] = "bar"
	return cr
}

func newSTSWithSpec(replicas int, pmp stsv1alpha1.PodManagementPolicyType) *stsv1alpha1.GameStatefulSet {
	sts := testutil.NewGameStatefulSet(replicas)
	sts.Spec.PodManagementPolicy = pmp
	return sts
}

func TestUpdateGameStatefulSet(t *testing.T) {
	stsFake.AddToScheme(scheme.Scheme)
	tests := []struct {
		name            string
		set             *stsv1alpha1.GameStatefulSet
		pods            []*corev1.Pod
		revisions       []*apps.ControllerRevision
		expectedError   error
		expectedActions []testing2.Action
	}{
		{
			name: "delete and recreate failed pods",
			set:  testutil.NewGameStatefulSet(3),
			pods: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodFailed, "foo-1", false, false),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(2, corev1.PodFailed, "foo-3", false, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-0"),
				expectCreatePodAction(newPodWithStatus(0, "", "foo-3", false, false)),
			},
		},
		{ // with parallel pod management, will create and delete pods and not wait for pods to be ready or complete
			// termination.
			name: "test update with parallel pod management",
			set:  newSTSWithSpec(3, stsv1alpha1.ParallelPodManagement),
			pods: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodFailed, "foo-1", false, false),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(2, corev1.PodFailed, "foo-3", false, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-0"),
				expectCreatePodAction(newPodWithStatus(0, "", "foo-3", false, false)),
				expectDeletePodAction(corev1.NamespaceDefault, "foo-2"),
				expectCreatePodAction(newPodWithStatus(2, "", "foo-3", false, false)),
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
				expectUpdatePodAction(corev1.NamespaceDefault, testutil.NewPod(1)),
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
				expectUpdatePodAction(corev1.NamespaceDefault, testutil.NewPod(1)),
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{ // TODO: Maybe this is a bug?
			// not specify the UpdateStrategy type, it should be OnDelete, but the spec default function isn't work.
			name: "test delete pod manually with default update strategy and parallel pod management policy",
			set: func() *stsv1alpha1.GameStatefulSet {
				sts := newSTSWithSpec(3, stsv1alpha1.ParallelPodManagement)
				sts.Spec.UpdateStrategy = stsv1alpha1.GameStatefulSetUpdateStrategy{}
				sts.Status.CurrentReplicas = 3
				return sts
			}(),
			pods: []*corev1.Pod{
				newPodWithStatus(1, corev1.PodRunning, "foo-2", true, false),
				newPodWithStatus(2, corev1.PodRunning, "foo-2", true, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
			expectedActions: []testing2.Action{
				expectCreatePodAction(newPodWithStatus(0, "", "foo-2", false, false)),
			},
		},
		{ // expect create new pod with new revision
			name: "test delete pod manually with OnDelete update strategy and parallel pod management policy",
			set: func() *stsv1alpha1.GameStatefulSet {
				sts := newSTSWithSpec(3, stsv1alpha1.ParallelPodManagement)
				sts.Spec.UpdateStrategy = stsv1alpha1.GameStatefulSetUpdateStrategy{
					Type: stsv1alpha1.OnDeleteGameStatefulSetStrategyType,
				}
				sts.Status.CurrentReplicas = 3
				return sts
			}(),
			pods: []*corev1.Pod{
				newPodWithStatus(1, corev1.PodRunning, "foo-2", true, false),
				newPodWithStatus(2, corev1.PodRunning, "foo-2", true, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
			expectedActions: []testing2.Action{
				expectCreatePodAction(newPodWithStatus(0, "", "foo-3", false, false)),
			},
		},
		{
			name: "force delete and recreate node lost pods",
			set:  testutil.NewGameStatefulSet(3),
			pods: []*corev1.Pod{
				func() *corev1.Pod {
					pod := newPodWithStatus(0, corev1.PodRunning, "foo-1", true, true)
					pod.Annotations[podNodeLostForceDeleteKey] = "false"
					return pod
				}(),
				newPodWithStatus(1, corev1.PodPending, "foo-1", false, true),
				newPodWithStatus(2, corev1.PodFailed, "foo-3", false, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
		},
		{
			name: "force delete and recreate node lost pods with parallel pod management",
			set:  newSTSWithSpec(3, stsv1alpha1.ParallelPodManagement),
			pods: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodPending, "foo-1", false, true),
				func() *corev1.Pod {
					pod := newPodWithStatus(1, corev1.PodRunning, "foo-1", true, true)
					pod.Annotations[podNodeLostForceDeleteKey] = "false"
					return pod
				}(),
				newPodWithStatus(2, corev1.PodFailed, "foo-3", false, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(3)),
				newCRWithName("foo-3", 3, testutil.NewGameStatefulSet(3)),
			},
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-0"),
				expectCreatePodAction(newPodWithStatus(0, "", "foo-3", false, false)),
				expectDeletePodAction(corev1.NamespaceDefault, "foo-2"),
				expectCreatePodAction(newPodWithStatus(2, "", "foo-3", false, false)),
			},
		},
		{
			name: "all pods are ready, delete condemned pods",
			set:  testutil.NewGameStatefulSet(1),
			pods: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-2", true, false),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(2, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(3, corev1.PodRunning, "foo-1", true, true),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
		},
		{
			name: "all pods are ready, delete condemned pods, with parallel pod management",
			set:  newSTSWithSpec(1, stsv1alpha1.ParallelPodManagement),
			pods: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-2", true, false),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(2, corev1.PodPending, "foo-1", false, false),
				newPodWithStatus(3, corev1.PodRunning, "foo-1", true, true),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-2"),
				expectDeletePodAction(corev1.NamespaceDefault, "foo-1"),
				expectDeletePodAction(corev1.NamespaceDefault, "foo-1"),
				expectDeletePodAction(corev1.NamespaceDefault, "foo-2"),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()
			stsClient := stsFake.NewSimpleClientset()
			hookClient := hookFake.NewSimpleClientset()

			kubeInformer := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
			stsInformer := stsInformers.NewSharedInformerFactory(stsClient, controller.NoResyncPeriodFunc())
			hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
			recorder := record.NewFakeRecorder(10)

			hookRunInformer := hookInformer.Tkex().V1alpha1().HookRuns()
			hookTemplateInformer := hookInformer.Tkex().V1alpha1().HookTemplates()
			preDeleteControl := predelete.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			preInplaceControl := preinplace.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			postInplaceControl := postinplace.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			ssc := NewDefaultGameStatefulSetControl(
				kubeClient,
				hookClient,
				NewRealGameStatefulSetPodControl(
					kubeClient,
					kubeInformer.Core().V1().Pods().Lister(),
					kubeInformer.Core().V1().PersistentVolumeClaims().Lister(),
					kubeInformer.Core().V1().Nodes().Lister(),
					recorder, newMetrics()),
				inplaceupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
				hotpatchupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
				NewRealGameStatefulSetStatusUpdater(stsClient, stsInformer.Tkex().V1alpha1().GameStatefulSets().Lister(),
					recorder),
				history.NewHistory(kubeClient, kubeInformer.Apps().V1().ControllerRevisions().Lister()),
				recorder,
				kubeInformer.Core().V1().Pods().Lister(),
				hookInformer.Tkex().V1alpha1().HookRuns().Lister(),
				hookInformer.Tkex().V1alpha1().HookTemplates().Lister(),
				preDeleteControl,
				preInplaceControl,
				postInplaceControl,
				newMetrics(),
			)

			stsClient.TkexV1alpha1().GameStatefulSets(corev1.NamespaceDefault).Create(context.TODO(), s.set, metav1.CreateOptions{})
			for _, revision := range s.revisions {
				kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().Add(revision)
			}
			for _, pod := range s.pods {
				kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
				kubeClient.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			}
			kubeInformer.Core().V1().Nodes().Informer().GetIndexer().Add(newNode(false))
			kubeClient.ClearActions()

			err := ssc.UpdateGameStatefulSet(s.set, s.pods)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			expectedActions := testutil.FilterActions(s.expectedActions, testutil.FilterUpdateAction, testutil.FilterOwnerRefer, testutil.FilterCreateAction)
			kubeActions := testutil.FilterActions(kubeClient.Actions(), testutil.FilterUpdateAction, testutil.FilterOwnerRefer, testutil.FilterCreateAction)
			if !testutil.EqualActions(expectedActions, kubeActions) {
				t.Errorf("expected actions \n\t%v\ngot \n\t%v", expectedActions, kubeActions)
			}
		})
	}
}

func newGameStatefulSet(replicas int, updateStrategy stsv1alpha1.GameStatefulSetUpdateStrategyType,
	hook string) *stsv1alpha1.GameStatefulSet {
	sts := testutil.NewGameStatefulSet(replicas)
	sts.Spec.UpdateStrategy.Type = updateStrategy
	if hook != "" {
		sts.Spec.PreInplaceUpdateStrategy.Hook = &v1alpha12.HookStep{
			TemplateName: hook,
		}
	}
	return sts
}

func setSTSPartition(partition int, set *stsv1alpha1.GameStatefulSet) *stsv1alpha1.GameStatefulSet {
	if set.Spec.UpdateStrategy.CanaryStrategy == nil {
		set.Spec.UpdateStrategy.CanaryStrategy = &stsv1alpha1.CanaryStrategy{}
	}
	if set.Spec.UpdateStrategy.CanaryStrategy.Steps == nil {
		set.Spec.UpdateStrategy.CanaryStrategy.Steps = []stsv1alpha1.CanaryStep{}
	}
	set.Spec.UpdateStrategy.CanaryStrategy.Steps = append(set.Spec.UpdateStrategy.CanaryStrategy.Steps,
		stsv1alpha1.CanaryStep{Partition: func() *int32 { i := int32(partition); return &i }()})
	return set
}

func newStatus() *stsv1alpha1.GameStatefulSetStatus {
	i := int32(0)
	return &stsv1alpha1.GameStatefulSetStatus{
		CollisionCount:   &i,
		CurrentStepIndex: &i,
	}
}

func newHookTemplate(name string) *v1alpha12.HookTemplate {
	ht := testutil.NewHookTemplate()
	ht.Name = name
	return ht
}

func TestHandleUpdateStrategy(t *testing.T) {
	stsFake.AddToScheme(scheme.Scheme)
	tests := []struct {
		name           string
		set            *stsv1alpha1.GameStatefulSet
		status         *stsv1alpha1.GameStatefulSetStatus
		revisions      []*apps.ControllerRevision
		updateRevision *apps.ControllerRevision
		replicas       []*corev1.Pod
		monotonic      bool
		// default pod list in the system
		podList             []*corev1.Pod
		hookTemplateList    []*v1alpha12.HookTemplate
		expectedError       error
		expectedSetStatus   *stsv1alpha1.GameStatefulSetStatus
		expectedKubeActions []testing2.Action
	}{
		{ // do noting
			name: "test OnDelete strategy",
			set:  newGameStatefulSet(1, stsv1alpha1.OnDeleteGameStatefulSetStrategyType, ""),
		},
		{
			name: "test InPlaceUpdate strategy",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.InplaceUpdateGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
				expectUpdatePodAction(corev1.NamespaceDefault, testutil.NewPod(1)),
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name: "test InPlaceUpdate strategy with empty revision",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.InplaceUpdateGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newControllerRevision("foo-1"),
				newControllerRevision("foo-2"),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			expectedError:  errors.New("but the diff not only contains replace operation of spec.containers[x].image"),
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
		},
		{
			name: "test InPlaceUpdate strategy without pod",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.InplaceUpdateGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			expectedError:  k8serrors.NewNotFound(corev1.SchemeGroupVersion.WithResource("pods").GroupResource(), "foo-1"),
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name: "inPlaceUpdate with hook",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.InplaceUpdateGameStatefulSetStrategyType, "hook1")),
			replicas: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-1", true, false),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, false),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			hookTemplateList: []*v1alpha12.HookTemplate{
				newHookTemplate("hook1"),
			},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.PreInplaceHookConditions = []v1alpha12.PreInplaceHookCondition{
					{
						PodName:   "foo-1",
						HookPhase: v1alpha12.HookPhasePending,
					},
				}
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectPatchPodAction(corev1.NamespaceDefault, "foo-1", types.StrategicMergePatchType),
			},
		},
		{
			name: "inPlaceUpdate with terminating pod",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.InplaceUpdateGameStatefulSetStrategyType, "hook1")),
			replicas: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-1", true, true),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, true),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			monotonic:      true,
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			hookTemplateList: []*v1alpha12.HookTemplate{
				newHookTemplate("hook1"),
			},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				return s
			}(),
		},
		{
			name: "test rollingUpdate strategy",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.RollingUpdateGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name: "rollingUpdate with terminating pod",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.RollingUpdateGameStatefulSetStrategyType, "hook1")),
			replicas: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-1", true, true),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, true),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			monotonic:      true,
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			hookTemplateList: []*v1alpha12.HookTemplate{
				newHookTemplate("hook1"),
			},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				return s
			}(),
		},
		{
			name: "test hotPatchUpdate strategy",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.HotPatchGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
				expectUpdatePodAction(corev1.NamespaceDefault, testutil.NewPod(1)),
			},
		},
		{
			name: "test hotPatchUpdate strategy with not pod",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.HotPatchGameStatefulSetStrategyType, "")),
			replicas: []*corev1.Pod{
				newPodWithControllerRevision("foo-1"),
				newPodWithControllerRevision("foo-1"),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(1)),
			status:         newStatus(),
			expectedError:  k8serrors.NewNotFound(corev1.SchemeGroupVersion.WithResource("pods").GroupResource(), "foo-1"),
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				s.CurrentReplicas = -1
				return s
			}(),
			expectedKubeActions: []testing2.Action{
				expectGetPodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name: "rollingUpdate with terminating pod",
			set:  setSTSPartition(1, newGameStatefulSet(2, stsv1alpha1.HotPatchGameStatefulSetStrategyType, "hook1")),
			replicas: []*corev1.Pod{
				newPodWithStatus(0, corev1.PodRunning, "foo-1", true, true),
				newPodWithStatus(1, corev1.PodRunning, "foo-1", true, true),
			},
			revisions: []*apps.ControllerRevision{
				newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
				newCRWithName("foo-2", 2, testutil.NewGameStatefulSet(1)),
			},
			updateRevision: newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(1)),
			monotonic:      true,
			status:         newStatus(),
			podList:        []*corev1.Pod{testutil.NewPod(1)},
			hookTemplateList: []*v1alpha12.HookTemplate{
				newHookTemplate("hook1"),
			},
			expectedSetStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				s := newStatus()
				return s
			}(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			kubeClient := fake.NewSimpleClientset()
			hookClient := hookFake.NewSimpleClientset()

			kubeInformer := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
			hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
			recorder := record.NewFakeRecorder(10)

			hookRunInformer := hookInformer.Tkex().V1alpha1().HookRuns()
			hookTemplateInformer := hookInformer.Tkex().V1alpha1().HookTemplates()
			preDeleteControl := predelete.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			preInplaceControl := preinplace.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			postInplaceControl := postinplace.New(kubeClient, hookClient, recorder,
				hookRunInformer.Lister(), hookTemplateInformer.Lister())
			inPlaceControl := inplaceupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey)

			for _, pod := range s.podList {
				kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
				kubeClient.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			}

			for _, template := range s.hookTemplateList {
				hookInformer.Tkex().V1alpha1().HookTemplates().Informer().GetIndexer().Add(template)
			}

			kubeClient.ClearActions()
			ssc := defaultGameStatefulSetControl{
				preInplaceControl:  preInplaceControl,
				preDeleteControl:   preDeleteControl,
				postInplaceControl: postInplaceControl,
				inPlaceControl:     inPlaceControl,
				kubeClient:         kubeClient,
				recorder:           recorder,
				podControl: NewRealGameStatefulSetPodControl(
					kubeClient,
					kubeInformer.Core().V1().Pods().Lister(),
					kubeInformer.Core().V1().PersistentVolumeClaims().Lister(),
					kubeInformer.Core().V1().Nodes().Lister(),
					recorder, newMetrics()),
				hotPatchControl: hotpatchupdate.NewForTypedClient(kubeClient, apps.ControllerRevisionHashLabelKey),
				metrics:         newMetrics(),
			}
			status, err := ssc.handleUpdateStrategy(s.set, s.status, s.revisions, s.updateRevision, s.replicas, s.monotonic)
			if !reflect.DeepEqual(s.expectedError, err) && !strings.Contains(err.Error(), s.expectedError.Error()) {
				t.Errorf("expected error\n\t%v\ngot \n\t%v", s.expectedError, err)
			}
			// ignore the time since it is always different
			testutil.FilterGameStatefulSetStatusTime(status)
			testutil.FilterGameStatefulSetStatusTime(s.expectedSetStatus)
			if !reflect.DeepEqual(status, s.expectedSetStatus) {
				t.Errorf("expected status\n\t%v\ngot\n\t%v", s.expectedSetStatus, status)
			}
			expectedActions := testutil.FilterActions(s.expectedKubeActions, testutil.FilterCreateAction,
				testutil.FilterUpdateAction, testutil.FilterPatchAction)
			kubeActions := testutil.FilterActions(kubeClient.Actions(), testutil.FilterCreateAction,
				testutil.FilterUpdateAction, testutil.FilterPatchAction)
			if !testutil.EqualActions(expectedActions, kubeActions) {
				t.Errorf("unexpected actions\n\t%v\nexpected\n\t%v", kubeActions, expectedActions)
			}
		})
	}
}

func TestTruncatePreInplaceHookConditions(t *testing.T) {
	tests := []struct {
		name           string
		pods           []*corev1.Pod
		newStatus      *stsv1alpha1.GameStatefulSetStatus
		expectedStatus *stsv1alpha1.GameStatefulSetStatus
	}{
		{
			name: "no pods",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			newStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreInplaceHookConditions = []v1alpha12.PreInplaceHookCondition{}
				return status
			}(),
			expectedStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreInplaceHookConditions = []v1alpha12.PreInplaceHookCondition{}
				return status
			}(),
		},
		{
			name: "truncate",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			newStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreInplaceHookConditions = []v1alpha12.PreInplaceHookCondition{
					{PodName: "foo-1"}, {PodName: "foo-0"}, {PodName: "foo-2"},
				}
				return status
			}(),
			expectedStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreInplaceHookConditions = []v1alpha12.PreInplaceHookCondition{
					{PodName: "foo-1"}, {PodName: "foo-0"},
				}
				return status
			}(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			ssc := &defaultGameStatefulSetControl{}
			ssc.truncatePreInplaceHookConditions(s.pods, s.newStatus)
			if !reflect.DeepEqual(s.expectedStatus, s.newStatus) {
				t.Errorf("expected status\n\t%v\ngot\n\t%v", s.expectedStatus, s.newStatus)
			}
		})
	}
}

func TestTruncatePreDeleteHookConditions(t *testing.T) {
	tests := []struct {
		name           string
		pods           []*corev1.Pod
		newStatus      *stsv1alpha1.GameStatefulSetStatus
		expectedStatus *stsv1alpha1.GameStatefulSetStatus
	}{
		{
			name: "no pods",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			newStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreDeleteHookConditions = []v1alpha12.PreDeleteHookCondition{}
				return status
			}(),
			expectedStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreDeleteHookConditions = []v1alpha12.PreDeleteHookCondition{}
				return status
			}(),
		},
		{
			name: "truncate",
			pods: []*corev1.Pod{
				testutil.NewPod(0),
				testutil.NewPod(1),
			},
			newStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreDeleteHookConditions = []v1alpha12.PreDeleteHookCondition{
					{PodName: "foo-1"}, {PodName: "foo-0"}, {PodName: "foo-2"},
				}
				return status
			}(),
			expectedStatus: func() *stsv1alpha1.GameStatefulSetStatus {
				status := newStatus()
				status.PreDeleteHookConditions = []v1alpha12.PreDeleteHookCondition{
					{PodName: "foo-1"}, {PodName: "foo-0"},
				}
				return status
			}(),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			ssc := &defaultGameStatefulSetControl{}
			ssc.truncatePreDeleteHookConditions(s.pods, s.newStatus)
			if !reflect.DeepEqual(s.expectedStatus, s.newStatus) {
				t.Errorf("expected status\n\t%v\ngot\n\t%v", s.expectedStatus, s.newStatus)
			}
		})
	}
}

func TestAdoptOrphanRevisions(t *testing.T) {
	tests := []struct {
		name           string
		set            *stsv1alpha1.GameStatefulSet
		revisions      []*apps.ControllerRevision
		existRevisions []*apps.ControllerRevision
		expectedError  error
	}{
		{
			name: "no revisions",
			set:  testutil.NewGameStatefulSet(1),
		},
		{
			name: "test adopt",
			set:  testutil.NewGameStatefulSet(2),
			revisions: []*apps.ControllerRevision{
				newControllerRevision("foo-1"),
				newControllerRevision("foo-2"),
			},
			existRevisions: []*apps.ControllerRevision{
				newControllerRevision("foo-1"),
				newControllerRevision("foo-2"),
			},
		},
		{
			name: "revisions not exist",
			set:  testutil.NewGameStatefulSet(2),
			revisions: []*apps.ControllerRevision{
				newControllerRevision("foo-1"),
				newControllerRevision("foo-2"),
			},
			existRevisions: []*apps.ControllerRevision{
				newControllerRevision("foo-1"),
			},
			expectedError: k8serrors.NewNotFound(apps.SchemeGroupVersion.WithResource("controllerrevisions").GroupResource(), "foo-2"),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			informer := informerFactory.Apps().V1().ControllerRevisions()
			for i := range s.existRevisions {
				informer.Informer().GetIndexer().Add(s.existRevisions[i])
			}
			controllerHistory := history.NewFakeHistory(informer)
			ssc := &defaultGameStatefulSetControl{
				controllerHistory: controllerHistory,
			}
			err := ssc.AdoptOrphanRevisions(s.set, s.revisions)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("expected error: %v, got: %v", s.expectedError, err)
			}
		})
	}
}
