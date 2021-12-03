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
	"encoding/json"
	"errors"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdfake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/fake"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	hookV1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookinformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"testing"
)

type testControl struct {
	Interface
	kubeClient    *fake.Clientset
	hookClient    *hookFake.Clientset
	hookInformers hookinformers.SharedInformerFactory
	kubeInformers informers.SharedInformerFactory
}

func newRealControl() testControl {
	// init kube controller
	kubeClient := fake.NewSimpleClientset()
	kubeInformers := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
	kubeStop := make(chan struct{})
	defer close(kubeStop)
	kubeInformers.Start(kubeStop)
	kubeInformers.WaitForCacheSync(kubeStop)

	// init hook controller
	hookClient := hookFake.NewSimpleClientset()
	hookFake.AddToScheme(scheme.Scheme)
	hookInformerFactory := hookinformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	hookStop := make(chan struct{})
	defer close(hookStop)
	hookInformerFactory.Start(hookStop)
	hookInformer := hookInformerFactory.Tkex().V1alpha1().HookRuns()
	hookTemplateInformer := hookInformerFactory.Tkex().V1alpha1().HookTemplates()
	hookInformerFactory.WaitForCacheSync(hookStop)
	return testControl{
		Interface: New(kubeClient, gdfake.NewSimpleClientset(), &record.FakeRecorder{}, expectations.NewScaleExpectations(),
			hookInformer.Lister(), hookTemplateInformer.Lister(),
			predelete.New(kubeClient, hookClient, &record.FakeRecorder{}, hookInformer.Lister(),
				hookTemplateInformer.Lister())),
		kubeClient:    kubeClient,
		hookClient:    hookClient,
		hookInformers: hookInformerFactory,
		kubeInformers: kubeInformers,
	}
}

func TestRealControl_ManageReplicasNil(t *testing.T) {
	control := newRealControl()
	updateDeploy := test.NewGameDeployment(1)
	updateDeploy.Spec.Replicas = nil
	manage, err := control.Manage(test.NewGameDeployment(1), test.NewGameDeployment(1), updateDeploy,
		"1", "2",
		[]*v1.Pod{test.NewPod(0)}, []*v1.Pod{test.NewPod(0)}, &test.NewGameDeployment(1).Status)
	if !reflect.DeepEqual(err, errors.New("spec.Replicas is nil")) {
		t.Errorf("err should be spec.Replicas is nil, but got %v", err)
	}
	if manage {
		t.Errorf("manage should be false")
	}
}

func TestValidateGameDeploymentPodIndex(t *testing.T) {
	control := newRealControl()
	deploy1 := test.NewGameDeployment(1)
	deploy1.Annotations = map[string]string{
		v1alpha1.GameDeploymentIndexOn: "true",
		"bcs.tencent.com/pod-index":    "1",
	}
	deploy2 := deploy1.DeepCopy()
	deploy2.Annotations[v1alpha1.GameDeploymentIndexRange] = "1-2"
	deploy3 := deploy1.DeepCopy()
	deploy3.Annotations[v1alpha1.GameDeploymentIndexRange] = `{"podStartIndex": 3, "podEndIndex": 2}`
	deploy4 := deploy1.DeepCopy()
	deploy4.Spec.Replicas = func() *int32 { a := int32(2); return &a }()
	deploy4.Annotations[v1alpha1.GameDeploymentIndexRange] = `{"podStartIndex": 1, "podEndIndex": 2}`
	tests := []struct {
		name          string
		deploy        *v1alpha1.GameDeployment
		exceptedError error
	}{
		{
			name:          "unset index range annotation",
			deploy:        deploy1,
			exceptedError: errors.New("gamedeployment foo inject index on, get index-range failed"),
		},
		{
			name:          "wrong index range annotation",
			deploy:        deploy2,
			exceptedError: errors.New("invalid character '-' after top-level value"),
		},
		{
			name:          "invalid index range",
			deploy:        deploy3,
			exceptedError: errors.New("gamedeployment foo invalid index range"),
		},
		{
			name:          "deploy scale replicas gt available indexs",
			deploy:        deploy4,
			exceptedError: errors.New("deploy foo scale replicas gt available indexs"),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			manage, err := control.Manage(s.deploy, s.deploy, s.deploy,
				"1", "2",
				[]*v1.Pod{test.NewPod(0)}, []*v1.Pod{test.NewPod(0)}, &s.deploy.Status)
			if manage {
				t.Errorf("manage should be false")
			}
			if err.Error() != s.exceptedError.Error() {
				t.Errorf("err should be %v, but got %v", s.exceptedError, err)
			}
		})
	}
}

func TestManagePods(t *testing.T) {
	// custom deploy1, used by case "have podToDelete and have hookTemplate, but PreDelete Hook not completed,
	//   so we can't delete pod"
	deploy1 := test.NewGameDeployment(2)
	deploy1.Spec.ScaleStrategy.PodsToDelete = []string{"foo-2"}
	deploy1.Spec.PreDeleteUpdateStrategy = v1alpha1.GameDeploymentPreDeleteUpdateStrategy{
		Hook: &hookV1alpha1.HookStep{
			TemplateName: "foo",
		}}
	tests := []struct {
		name                string
		currentDeploy       *v1alpha1.GameDeployment
		updateDeploy        *v1alpha1.GameDeployment
		currentRevision     string
		updateRevision      string
		pods                []*v1.Pod
		hookTemplates       []*hookV1alpha1.HookTemplate
		hookRuns            []*hookV1alpha1.HookRun
		newStatus           *v1alpha1.GameDeploymentStatus
		exceptedScaling     bool
		exceptedError       error
		exceptedKubeActions []clientTesting.Action
		exceptedHookActions []clientTesting.Action
	}{
		// with pod to delete
		{
			name:          "have podToDelete and can delete",
			currentDeploy: test.NewGameDeployment(1),
			updateDeploy: func() *v1alpha1.GameDeployment {
				d := test.NewGameDeployment(2)
				d.Spec.ScaleStrategy.PodsToDelete = []string{"foo-1", "foo-2"}
				return d
			}(),
			currentRevision: "1",
			updateRevision:  "2",
			pods: []*v1.Pod{
				test.NewPod(0),
				test.NewPod(1),
				func() *v1.Pod {
					pod := test.NewPod(2)
					pod.Status.Phase = v1.PodFailed
					return pod
				}(),
			},
			newStatus:       &v1alpha1.GameDeploymentStatus{},
			exceptedScaling: true,
			exceptedError:   nil,
			exceptedKubeActions: []clientTesting.Action{
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
				),
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-2",
				),
			},
			exceptedHookActions: []clientTesting.Action{},
		},
		{
			name:            "have podToDelete and have hookTemplate, but hookTemplates list are empty",
			currentDeploy:   test.NewGameDeployment(1),
			updateDeploy:    deploy1,
			currentRevision: "1",
			updateRevision:  "2",
			pods: []*v1.Pod{
				test.NewPod(0),
				test.NewPod(1),
				func() *v1.Pod {
					pod := test.NewPod(2)
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookRuns:        []*hookV1alpha1.HookRun{},
			newStatus:       &v1alpha1.GameDeploymentStatus{},
			exceptedScaling: false,
			exceptedError: apierrors.NewNotFound(schema.GroupResource{Group: "tkex.tencent.com", Resource: "hooktemplate"},
				"foo"),
			exceptedKubeActions: []clientTesting.Action{},
			exceptedHookActions: []clientTesting.Action{},
		},
		{
			name:            "have podToDelete and have hookTemplate, but PreDelete Hook not completed, so we can't delete pod",
			currentDeploy:   test.NewGameDeployment(1),
			updateDeploy:    deploy1,
			currentRevision: "1",
			updateRevision:  "2",
			pods: []*v1.Pod{
				test.NewPod(0),
				test.NewPod(1),
				func() *v1.Pod {
					pod := test.NewPod(2)
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates: []*hookV1alpha1.HookTemplate{
				test.NewHookTemplate(),
			},
			newStatus:       &v1alpha1.GameDeploymentStatus{},
			exceptedScaling: false,
			exceptedError:   nil,
			exceptedKubeActions: []clientTesting.Action{
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-2",
					types.StrategicMergePatchType,
					func() []byte {
						currentAnnotations := map[string]string{
							predelete.DeletingAnnotation: "true",
						}
						patchData := map[string]interface{}{
							"metadata": map[string]map[string]string{
								"annotations": currentAnnotations,
							},
						}
						playLoadBytes, _ := json.Marshal(patchData)
						return playLoadBytes
					}(),
				),
			},
			exceptedHookActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
			},
		},
		// without pod to delete
		{
			name:            "scale do nothing",
			currentDeploy:   test.NewGameDeployment(3),
			updateDeploy:    test.NewGameDeployment(3),
			currentRevision: "1",
			updateRevision:  "1",
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(2)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates:       []*hookV1alpha1.HookTemplate{},
			newStatus:           &v1alpha1.GameDeploymentStatus{},
			exceptedScaling:     false,
			exceptedError:       nil,
			exceptedKubeActions: []clientTesting.Action{},
			exceptedHookActions: []clientTesting.Action{},
		},
		{
			name:            "scale out",
			currentDeploy:   test.NewGameDeployment(1),
			updateDeploy:    test.NewGameDeployment(2),
			currentRevision: "1",
			updateRevision:  "2",
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates:   []*hookV1alpha1.HookTemplate{},
			newStatus:       &v1alpha1.GameDeploymentStatus{},
			exceptedScaling: true,
			exceptedError:   nil,
			exceptedKubeActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					nil,
				),
			},
			exceptedHookActions: []clientTesting.Action{},
		},
		{
			name: "scale in",
			currentDeploy: func() *v1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Annotations = map[string]string{
					v1alpha1.GameDeploymentIndexOn:    "true",
					v1alpha1.GameDeploymentIndexRange: `{"podStartIndex": 0, "podEndIndex": 1000}`,
				}
				return d
			}(),
			updateDeploy: func() *v1alpha1.GameDeployment {
				d := test.NewGameDeployment(1)
				d.Annotations = map[string]string{
					v1alpha1.GameDeploymentIndexOn:    "true",
					v1alpha1.GameDeploymentIndexRange: `{"podStartIndex": 0, "podEndIndex": 1000}`,
				}
				return d
			}(),
			currentRevision: "1",
			updateRevision:  "2",
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Annotations = map[string]string{
						v1alpha1.GameDeploymentIndexID: "0",
					}
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "2"
					pod.Labels[v1alpha1.GameDeploymentInstanceID] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(2)
					pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "1"
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates:   []*hookV1alpha1.HookTemplate{},
			newStatus:       &v1alpha1.GameDeploymentStatus{},
			exceptedScaling: true,
			exceptedError:   nil,
			exceptedKubeActions: []clientTesting.Action{
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
				),
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-2",
				),
			},
			exceptedHookActions: []clientTesting.Action{},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			control := newRealControl()
			// mock pods objects
			for i := range s.pods {
				_, _ = control.kubeClient.CoreV1().Pods(v1.NamespaceDefault).Create(s.pods[i])
				_ = control.kubeInformers.Core().V1().Pods().Informer().GetIndexer().Add(s.pods[i])
			}
			// mock hookTemplates objects
			for _, template := range s.hookTemplates {
				_, _ = control.hookClient.TkexV1alpha1().HookTemplates(v1.NamespaceDefault).Create(template)
				_ = control.hookInformers.Tkex().V1alpha1().HookTemplates().Informer().GetIndexer().Add(template)
			}
			// mock hookRuns objects
			for _, hr := range s.hookRuns {
				_, _ = control.hookClient.TkexV1alpha1().HookRuns(v1.NamespaceDefault).Create(hr)
				_ = control.hookInformers.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr)
			}
			// clear test data
			control.kubeClient.ClearActions()
			control.hookClient.ClearActions()

			scaling, err := control.Manage(s.currentDeploy, s.currentDeploy, s.updateDeploy,
				s.currentRevision, s.updateRevision, s.pods, s.pods, s.newStatus)
			if scaling != s.exceptedScaling {
				t.Errorf("scaling should be %v, but got %v", s.exceptedScaling, scaling)
			}
			if !reflect.DeepEqual(err, s.exceptedError) {
				t.Errorf("err should be %v, but got %v", s.exceptedError, err)
			}
			// only compare verb, version, resources, namespace, exclude object, because pod's name is random string
			kubeActions := test.FilterActionsObject(control.kubeClient.Actions())
			hookActions := test.FilterActionsObject(control.hookClient.Actions())
			if !test.EqualActions(s.exceptedKubeActions, kubeActions) {
				t.Errorf("kube actions should be %v, but got %v", s.exceptedKubeActions, kubeActions)
			}
			if !test.EqualActions(s.exceptedHookActions, hookActions) {
				t.Errorf("hook actions should be %v, but got %v", s.exceptedHookActions, control.hookClient.Actions())
			}
		})
	}
}
