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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	gdcore "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/core"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientTesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"strings"
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
	kubeInformers.Start(kubeStop)
	kubeInformers.WaitForCacheSync(kubeStop)

	// init hook controller
	hookClient := hookFake.NewSimpleClientset()
	_ = hookFake.AddToScheme(scheme.Scheme)
	hookInformerFactory := hookinformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	hookStop := make(chan struct{})
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
			preDeleteControl, preInplaceControl, postInpalceControl, gdmetrics.NewMetrics()),
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
		expectedRequeueDuration time.Duration
		expectedErrorFn         func(got error) (expected bool, want error)
		expectedKubeActions     []clientTesting.Action
		expectedHookActions     []clientTesting.Action
	}{
		{
			name: "update success",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{Type: gdv1alpha1.RollingGameDeploymentUpdateStrategyType}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				set := test.NewGameDeployment(3)
				set.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(set, 2, set.Status.CollisionCount)
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
			expectedRequeueDuration: 0,
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
				),
				clientTesting.NewDeleteAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
				),
			},
		},
		{
			name: "update with preDelete hook, but not completed",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{Type: gdv1alpha1.RollingGameDeploymentUpdateStrategyType}
				d.Spec.PreDeleteUpdateStrategy = gdv1alpha1.GameDeploymentPreDeleteUpdateStrategy{
					Hook: &hookV1alpha1.HookStep{
						TemplateName: "foo",
					}}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				set := test.NewGameDeployment(3)
				set.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(set, 2, set.Status.CollisionCount)
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
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "2",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates: []*hookV1alpha1.HookTemplate{
				test.NewHookTemplate(),
			},
			expectedRequeueDuration: 0,
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
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
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
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
			expectedHookActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
			},
		},
		{
			name: "update with preDelete hook, test existingRun",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{Type: gdv1alpha1.RollingGameDeploymentUpdateStrategyType}
				d.Spec.PreDeleteUpdateStrategy = gdv1alpha1.GameDeploymentPreDeleteUpdateStrategy{
					Hook: &hookV1alpha1.HookStep{
						TemplateName: "foo",
					}}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				set := test.NewGameDeployment(3)
				set.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(set, 2, set.Status.CollisionCount)
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
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "1",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates: []*hookV1alpha1.HookTemplate{
				test.NewHookTemplate(),
			},
			expectedRequeueDuration: 0,
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
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
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
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
			// because hook run create before, newly hook run will do 'GET' action first, then create again with another name
			expectedHookActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
				clientTesting.NewGetAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					"predelete-1--foo",
				),
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
			},
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
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedRequeueDuration: 0,
		},
		{
			name: "inPlace update success",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(1)
					for i := range d.Spec.Template.Spec.Containers {
						d.Spec.Template.Spec.Containers[i].Image += "+"
					}
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedRequeueDuration: 10 * time.Second,
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewGetAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
				),
				clientTesting.NewUpdateAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					nil,
				),
				clientTesting.NewGetAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
				),
			},
		},
		{
			name: "inPlace update with wrong spec",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(2)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					// update type isn't set
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(2)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				want := fmt.Errorf("invalid update strategy type")
				return reflect.DeepEqual(err, want), want
			},
			expectedRequeueDuration: 0,
		},
		{
			name: "inPlace update spec diff not only contains image",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(3)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(1)
					// set different template
					d.Spec.Template.Spec.NodeName = "test"
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Spec.NodeName = "test"
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				want := "but the diff not only contains replace operation of spec.containers[x].image"
				return strings.Contains(err.Error(), want), errors.New(want)
			},
			expectedRequeueDuration: 0,
		},
		{
			name: "inPlace update with preInPlace update strategy",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.PreInplaceUpdateStrategy = gdv1alpha1.GameDeploymentPreInplaceUpdateStrategy{
					Hook: &hookV1alpha1.HookStep{
						TemplateName: "foo",
					},
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(1)
					for i := range d.Spec.Template.Spec.Containers {
						d.Spec.Template.Spec.Containers[i].Image += "+"
					}
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
			},
			hookTemplates: []*hookV1alpha1.HookTemplate{
				test.NewHookTemplate(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedRequeueDuration: 0,
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
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
			expectedHookActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
			},
		},
		{
			name: "inPlace update with preInPlace update strategy, and pod isn't running",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(2)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.PreInplaceUpdateStrategy = gdv1alpha1.GameDeploymentPreInplaceUpdateStrategy{
					Hook: &hookV1alpha1.HookStep{
						TemplateName: "foo",
					},
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(2)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(2)
					for i := range d.Spec.Template.Spec.Containers {
						d.Spec.Template.Spec.Containers[i].Image += "+"
					}
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					pod.Status.Phase = v1.PodRunning
					return pod
				}(),
				func() *v1.Pod {
					pod := test.NewPod(1)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			hookTemplates: []*hookV1alpha1.HookTemplate{
				test.NewHookTemplate(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedRequeueDuration: 10 * time.Second,
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewGetAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
				),
				clientTesting.NewUpdateAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					nil,
				),
				clientTesting.NewGetAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-1",
				),
				clientTesting.NewPatchAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
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
			expectedHookActions: []clientTesting.Action{
				clientTesting.NewCreateAction(
					schema.GroupVersion{Group: "tkex.tencent.com", Version: "v1alpha1"}.WithResource("hookruns"),
					v1.NamespaceDefault,
					nil,
				),
			},
		},
		{
			name: "hotPatch update success",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(1)
				for i := range d.Spec.Template.Spec.Containers {
					d.Spec.Template.Spec.Containers[i].Image += "+"
				}
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(1)
					for i := range d.Spec.Template.Spec.Containers {
						d.Spec.Template.Spec.Containers[i].Image += "+"
					}
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				return err == nil, nil
			},
			expectedRequeueDuration: 0,
			expectedKubeActions: []clientTesting.Action{
				clientTesting.NewGetAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					"foo-0",
				),
				clientTesting.NewUpdateAction(
					v1.SchemeGroupVersion.WithResource("pods"),
					v1.NamespaceDefault,
					nil,
				),
			},
		},
		{
			name: "hotPatch update error",
			updateDeploy: func() *gdv1alpha1.GameDeployment {
				d := test.NewGameDeployment(1)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				return d
			}(),
			updateRevision: func() *apps.ControllerRevision {
				control := revision.NewRevisionControl()
				d := test.NewGameDeployment(1)
				d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
					Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
					InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
				}
				d.Status.CollisionCount = new(int32)
				currentRevision, err := control.NewRevision(d, 2, d.Status.CollisionCount)
				if err != nil {
					t.Fatalf("create revision error: %v", err)
				}
				return currentRevision
			}(),
			revisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					control := revision.NewRevisionControl()
					d := test.NewGameDeployment(1)
					// set different template
					d.Spec.Template.Spec.NodeName = "test"
					d.Spec.UpdateStrategy = gdv1alpha1.GameDeploymentUpdateStrategy{
						Type:                  gdv1alpha1.HotPatchGameDeploymentUpdateStrategyType,
						InPlaceUpdateStrategy: &inplaceupdate.InPlaceUpdateStrategy{GracePeriodSeconds: int32(10)},
					}
					d.Status.CollisionCount = new(int32)
					currentRevision, _ := control.NewRevision(d, 1, d.Status.CollisionCount)
					currentRevision.Name = "foo-1-1"
					return currentRevision
				}(),
			},
			pods: []*v1.Pod{
				func() *v1.Pod {
					pod := test.NewPod(0)
					pod.Spec.NodeName = "test"
					pod.Labels = map[string]string{
						apps.ControllerRevisionHashLabelKey: "foo-1-1",
					}
					return pod
				}(),
			},
			expectedErrorFn: func(err error) (bool, error) {
				want := "but the diff not only contains replace operation of spec.containers[x].image"
				return strings.Contains(err.Error(), want), errors.New(want)
			},
			expectedRequeueDuration: 0,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			control := newRealControl()
			// mock pods objects
			for i := range s.pods {
				_, _ = control.kubeClient.CoreV1().Pods(v1.NamespaceDefault).Create(context.TODO(), s.pods[i], metav1.CreateOptions{})
				_ = control.kubeInformers.Core().V1().Pods().Informer().GetIndexer().Add(s.pods[i])
			}
			// mock hookTemplates objects
			for _, template := range s.hookTemplates {
				_, _ = control.hookClient.TkexV1alpha1().HookTemplates(v1.NamespaceDefault).Create(context.TODO(), template, metav1.CreateOptions{})
				_ = control.hookInformers.Tkex().V1alpha1().HookTemplates().Informer().GetIndexer().Add(template)
			}
			// mock hookRuns objects
			for _, hr := range s.hookRuns {
				_, _ = control.hookClient.TkexV1alpha1().HookRuns(v1.NamespaceDefault).Create(context.TODO(), hr, metav1.CreateOptions{})
				_ = control.hookInformers.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr)
			}
			// clear test data
			control.kubeClient.ClearActions()
			control.hookClient.ClearActions()

			requeueDuration, err := control.Manage(test.NewGameDeployment(1), s.updateDeploy, s.updateRevision,
				s.revisions, s.pods, &test.NewGameDeployment(1).Status)
			kubeActions := test.FilterActions(control.kubeClient.Actions(), test.FilterCreateAction, test.FilterUpdateAction)
			hookActions := test.FilterActions(control.hookClient.Actions(), test.FilterCreateAction, test.FilterUpdateAction)
			if expected, want := s.expectedErrorFn(err); !expected {
				t.Errorf("got error %v, want %v", err, want)
			}
			if s.expectedRequeueDuration != requeueDuration {
				t.Errorf("got requeueDuration %v, want %v", requeueDuration, s.expectedRequeueDuration)
			}
			if !test.EqualActions(s.expectedKubeActions, kubeActions) {
				t.Errorf("kube actions should be\n\t%v,\ngot:\n\t%v", s.expectedKubeActions, kubeActions)
			}
			if !test.EqualActions(s.expectedHookActions, hookActions) {
				t.Errorf("hook actions should be\n\t%v,\ngot:\n\t%v", s.expectedHookActions, hookActions)
			}
		})
	}
}

func getInt32Pointer(i int32) *int32 {
	return &i
}

func TestCalculateUpdateCount(t *testing.T) {
	readyPod := func() *v1.Pod {
		return &v1.Pod{Status: v1.PodStatus{Phase: v1.PodRunning, Conditions: []v1.PodCondition{{Type: v1.PodReady, Status: v1.ConditionTrue}}}}
	}
	cases := []struct {
		name              string
		strategy          gdv1alpha1.GameDeploymentUpdateStrategy
		totalReplicas     int
		waitUpdateIndexes []int
		pods              []*v1.Pod
		expectedResult    int
	}{
		{
			name:              "1",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			name:              "2",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), {}, readyPod()},
			expectedResult:    0,
		},
		{
			name:              "3",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			name:              "4",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{},
			totalReplicas:     10,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			pods:              []*v1.Pod{{}, readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), {}, readyPod()},
			expectedResult:    1,
		},
		{
			name:              "5",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(3))},
			totalReplicas:     3,
			waitUpdateIndexes: []int{0, 1},
			pods:              []*v1.Pod{{}, readyPod(), readyPod()},
			expectedResult:    2,
		},
		{
			name:              "6",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromString("50%"))},
			totalReplicas:     8,
			waitUpdateIndexes: []int{0, 1, 2, 3, 4, 5, 6},
			pods:              []*v1.Pod{{}, readyPod(), {}, readyPod(), readyPod(), readyPod(), readyPod(), {}},
			expectedResult:    3,
		},
		{
			// maxUnavailable = 0 and maxSurge = 2, usedSurge = 1
			name: "7",
			strategy: gdv1alpha1.GameDeploymentUpdateStrategy{
				MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(0)),
				MaxSurge:       intstrutil.ValueOrDefault(nil, intstrutil.FromInt(2)),
			},
			totalReplicas:     4,
			waitUpdateIndexes: []int{0, 1},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod(), readyPod(), readyPod()},
			expectedResult:    1,
		},
		{
			// maxUnavailable = 1 and maxSurge = 2, usedSurge = 2
			name: "8",
			strategy: gdv1alpha1.GameDeploymentUpdateStrategy{
				MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromInt(1)),
				MaxSurge:       intstrutil.ValueOrDefault(nil, intstrutil.FromInt(2)),
			},
			totalReplicas:     4,
			waitUpdateIndexes: []int{0, 1, 2},
			pods:              []*v1.Pod{readyPod(), readyPod(), readyPod(), readyPod(), readyPod(), readyPod()},
			expectedResult:    3,
		},
		{
			// wait update index <= current partition
			name:              "9",
			strategy:          gdv1alpha1.GameDeploymentUpdateStrategy{Partition: getInt32Pointer(2), MaxUnavailable: intstrutil.ValueOrDefault(nil, intstrutil.FromString("50%"))},
			totalReplicas:     8,
			waitUpdateIndexes: []int{},
			pods:              []*v1.Pod{{}, readyPod(), {}, readyPod(), readyPod(), readyPod(), readyPod(), {}},
			expectedResult:    0,
		},
	}
	coreControl := gdcore.New(&gdv1alpha1.GameDeployment{})
	for _, s := range cases {
		t.Run(s.name, func(t *testing.T) {
			res := calculateUpdateCount(&gdv1alpha1.GameDeployment{}, coreControl, s.strategy, 0,
				s.totalReplicas, s.waitUpdateIndexes, s.pods)
			if res != s.expectedResult {
				t.Fatalf("expected %d, got %d", s.expectedResult, res)
			}
		})
	}
}
