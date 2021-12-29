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

package update

import (
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/revision"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	hookV1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookinformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/postinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/predelete"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/preinplace"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"testing"
	"time"
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
	hookRunInformer := hookInformerFactory.Tkex().V1alpha1().HookRuns()
	hookTemplateInformer := hookInformerFactory.Tkex().V1alpha1().HookTemplates()
	hookInformerFactory.WaitForCacheSync(hookStop)

	recorder := &record.FakeRecorder{}
	preDeleteControl := predelete.New(kubeClient, hookClient, recorder, hookRunInformer.Lister(), hookTemplateInformer.Lister())
	preInplaceControl := preinplace.New(kubeClient, hookClient, recorder, hookRunInformer.Lister(), hookTemplateInformer.Lister())
	postInpalceControl := postinplace.New(kubeClient, hookClient, recorder,
		hookRunInformer.Lister(), hookTemplateInformer.Lister())
	return testControl{
		Interface: New(kubeClient, recorder, expectations.NewScaleExpectations(),
			expectations.NewUpdateExpectations(util.GetPodRevision),
			hookRunInformer.Lister(), hookTemplateInformer.Lister(),
			preDeleteControl, preInplaceControl, postInpalceControl),
		kubeClient:    kubeClient,
		hookClient:    hookClient,
		hookInformers: hookInformerFactory,
		kubeInformers: kubeInformers,
	}
}

func TestGameDeploymentUpdateManage(t *testing.T) {
	_ = gdscheme.AddToScheme(scheme.Scheme)
	tests := []struct {
		name                    string
		updateDeploy            *gdv1alpha1.GameDeployment
		updateRevision          *apps.ControllerRevision
		revisions               []*apps.ControllerRevision
		pods                    []*v1.Pod
		hookTemplates           []*hookV1alpha1.HookTemplate
		hookRuns                []*hookV1alpha1.HookRun
		exceptedRequeueDuration time.Duration
		exceptedError           error
		exceptedKubeActions     []clientTesting.Action
		exceptedHookActions     []clientTesting.Action
	}{
		{
			name: "update success",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{Type: gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				set := test.NewGameDeployment(3)
				set.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(set, 3, set.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					set := test.NewGameDeployment(1)
					set.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(set, 1, set.Status.CollisionCount)
					return currentRevision
				}(),
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					set := test.NewGameDeployment(2)
					set.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(set, 2, set.Status.CollisionCount)
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					return pod
				}(),
			},
			exceptedRequeueDuration: 0,
			exceptedError:           nil,
		},
		{
			name: "paused",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: true,
				}
				return d
			}(),
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
		},
		{
			name: "inPlace update alright",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				set := test.NewGameDeployment(3)
				set.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(set, 3, set.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					set := test.NewGameDeployment(1)
					set.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(set, 1, set.Status.CollisionCount)
					return currentRevision
				}(),
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					set := test.NewGameDeployment(2)
					set.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(set, 2, set.Status.CollisionCount)
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					return pod
				}(),
			},
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
		},
		{
			name: "inPlace update with wrong spec",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: true,
				}
				return d
			}(),
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
		},
		{
			name: "rolling update alright",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: true,
				}
				return d
			}(),
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
		},
		{
			name: "rolling update alright",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: true,
				}
				return d
			}(),
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
		},
		{
			name: "hotPatch update alright",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: true,
				}
				return d
			}(),
			exceptedError:           nil,
			exceptedRequeueDuration: 0,
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
			control.kubeClient.ClearActions()
			requeueDuration, err := control.Manage(test.NewGameDeployment(1), s.updateDeploy, s.updateRevision,
				s.revisions, s.pods, &test.NewGameDeployment(1).Status)
			if !reflect.DeepEqual(err, s.exceptedError) {
				t.Errorf("got error %v, want %v", err, s.exceptedError)
			}
			if s.exceptedRequeueDuration != requeueDuration {
				t.Errorf("got requeueDuration %v, want %v", requeueDuration, s.exceptedRequeueDuration)
			}
			if !test.EqualActions(s.exceptedKubeActions, control.kubeClient.Actions()) {
				t.Errorf("kube actions should be %v, but got %v", s.exceptedKubeActions, control.kubeClient.Actions())
			}
			if !test.EqualActions(s.exceptedHookActions, control.hookClient.Actions()) {
				t.Errorf("hook actions should be %v, but got %v", s.exceptedHookActions, control.hookClient.Actions())
			}
		})
	}
}
