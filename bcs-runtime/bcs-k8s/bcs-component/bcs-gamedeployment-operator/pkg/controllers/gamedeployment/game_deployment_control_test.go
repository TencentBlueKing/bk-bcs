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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/revision"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	v1alpha12 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeSchema "k8s.io/apimachinery/pkg/runtime/schema"
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
	"testing"
	"time"
)

func TestGetActiveRevisions(t *testing.T) {
	_ = gdscheme.AddToScheme(scheme.Scheme)
	revisionControl := revision.NewRevisionControl()
	var collisionCount int32

	// initialize test data
	deploy1 := test.NewGameDeployment(1)
	// because revision will hash the spec.template, so we need to change the spec.template
	deploy1.Spec.Template.Labels["test"] = "test1"
	dRev1, err := revisionControl.NewRevision(deploy1, 1, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	deploy2 := test.NewGameDeployment(2)
	deploy2.Spec.Template.Labels["test"] = "test2"
	dRev2, err := revisionControl.NewRevision(deploy2, 2, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	deploy3 := test.NewGameDeployment(3)
	deploy3.Spec.Template.Labels["test"] = "test3"
	dRev3, err := revisionControl.NewRevision(deploy3, 3, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	dRev4, err := revisionControl.NewRevision(deploy2, 4, &collisionCount)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		deploy       *v1alpha1.GameDeployment
		revisions    []*apps.ControllerRevision
		podRevisions sets.String

		expectedCurrentRevision *apps.ControllerRevision
		expectedUpdateRevision  *apps.ControllerRevision
		expectedCollisionCount  int32
		expectedError           error
	}{
		{ // the equivalent revision is the latest revision
			name:                    "the equivalent revision is the latest revision",
			deploy:                  deploy3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev2.Name: {}},
			expectedCurrentRevision: dRev2,
			expectedUpdateRevision:  dRev3,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // the equivalent revision isn't the latest revision
			name:                    "the equivalent revision isn't the latest revision",
			deploy:                  deploy2,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2, dRev3},
			podRevisions:            map[string]sets.Empty{dRev3.Name: {}},
			expectedCurrentRevision: dRev3,
			expectedUpdateRevision:  dRev4,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // haven't equivalent revision
			name:                    "haven't equivalent revision",
			deploy:                  deploy3,
			revisions:               []*apps.ControllerRevision{dRev1, dRev2},
			podRevisions:            map[string]sets.Empty{dRev1.Name: {}},
			expectedCurrentRevision: dRev1,
			expectedUpdateRevision:  dRev3,
			expectedCollisionCount:  0,
			expectedError:           nil,
		},
		{ // when initializing, the latest revision is the current revision
			name:                    "when initializing",
			deploy:                  deploy3,
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
				_ = informer.Informer().GetIndexer().Add(s.revisions[i])
			}
			controllerHistory := history.NewFakeHistory(informer)
			control := &defaultGameDeploymentControl{revisionControl: revisionControl, controllerHistory: controllerHistory}

			currentRevision, updateRevision, collisionCount, err := control.getActiveRevisions(s.deploy, s.revisions, s.podRevisions)
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

func newHR(name string, phase v1alpha12.HookPhase, deleted bool, hrType string) *v1alpha12.HookRun {
	hr := test.NewHookRunFromTemplate(test.NewHookTemplate(), test.NewGameDeployment(1))
	hr.Name = name
	hr.Status.Phase = phase
	if deleted {
		hr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	}
	if hrType == "" {
		hrType = commonhookutil.HookRunTypePreDeleteLabel
	}
	hr.Labels[commonhookutil.HookRunTypeLabel] = hrType
	return hr
}

func TestDeleteUnexpectedPreDeleteHRs(t *testing.T) {
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
			gdc := &defaultGameDeploymentControl{
				hookClient: hookClient,
			}
			err := gdc.deleteUnexpectedPreDeleteHookRuns(s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !test.EqualActions(s.expectedActions, hookClient.Actions()) {
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
				test.NewPod(0),
				test.NewPod(1),
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
					pod := test.NewPod(0)
					pod.Labels[apps.ControllerRevisionHashLabelKey] = "1"
					pod.Labels[v1alpha1.GameDeploymentInstanceID] = "1"
					return pod
				}(),
				test.NewPod(1),
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
			gdc := &defaultGameDeploymentControl{
				hookClient: hookClient,
			}
			err := gdc.truncatePreDeleteHookRuns(test.NewGameDeployment(1), s.pods, s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !test.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestDeleteUnexpectedPreInPlaceHRs(t *testing.T) {
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
			gdc := &defaultGameDeploymentControl{
				hookClient: hookClient,
			}
			err := gdc.deleteUnexpectedPreInplaceHookRuns(s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !test.EqualActions(s.expectedActions, hookClient.Actions()) {
				t.Errorf("expected actions %v, got %v", s.expectedActions, hookClient.Actions())
			}
		})
	}
}

func TestTruncatePreInPlaceHookRuns(t *testing.T) {
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
				test.NewPod(0),
				test.NewPod(1),
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
					pod := test.NewPod(0)
					pod.Labels[apps.ControllerRevisionHashLabelKey] = "1"
					pod.Labels[v1alpha1.GameDeploymentInstanceID] = "1"
					return pod
				}(),
				test.NewPod(1),
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
			gdc := &defaultGameDeploymentControl{
				hookClient: hookClient,
			}
			err := gdc.truncatePreInplaceHookRuns(test.NewGameDeployment(1), s.pods, s.hrList)
			if err != s.expectedError {
				t.Errorf("expected error %v, got %v", s.expectedError, err)
			}
			if !test.EqualActions(s.expectedActions, hookClient.Actions()) {
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

	gdc := &defaultGameDeploymentControl{
		kubeClient: kubeClient,
		recorder:   recorder,
		podLister:  kubeInformer.Core().V1().Pods().Lister(),
		predeleteControl: predelete.New(kubeClient, hookClient, recorder,
			hookInformer.Tkex().V1alpha1().HookRuns().Lister(), hookInformer.Tkex().V1alpha1().HookTemplates().Lister()),
		metrics: gdmetrics.NewMetrics(),
	}

	pod := test.NewPod(0)
	_ = kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
	_ = gdc.deletePod(test.NewGameDeployment(1), pod.Name, &v1alpha1.GameDeploymentStatus{})
	if got, want := len(kubeClient.Actions()), 1; got != want {
		t.Fatalf("not expected pod actions count, want: %d, got: %d", want, got)
	}
	if !kubeClient.Actions()[0].Matches("delete", "pods") {
		t.Errorf("not expected pod actions verb")
	}
	kubeClient.ClearActions()

	// test pod not exist
	pod2 := test.NewPod(2)
	_ = gdc.deletePod(test.NewGameDeployment(1), pod2.Name, &v1alpha1.GameDeploymentStatus{})
	if got, want := len(kubeClient.Actions()), 0; got != want {
		t.Fatalf("not expected pod actions count, want: %d, got: %d", want, got)
	}
}

func newDeploy(limit int32) *v1alpha1.GameDeployment {
	d := test.NewGameDeployment(1)
	d.Spec.RevisionHistoryLimit = &limit
	return d
}

func newPodWithControllerRevision(revision string) *corev1.Pod {
	pod := test.NewPod(1)
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
			gdc := &defaultGameDeploymentControl{
				controllerHistory: history.NewFakeHistory(kubeInformer.Apps().V1().ControllerRevisions()),
			}
			for _, controllerRevision := range s.revisions {
				_ = kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().Add(controllerRevision)
			}
			err := gdc.truncateHistory(newDeploy(s.revisionHistoryLimit), s.pods, s.revisions,
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
