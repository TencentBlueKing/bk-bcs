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

package gamedeployment

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	v1alpha12 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"testing"
)

func TestGetHookRunFromGameDeployment(t *testing.T) {
	hookClient := hookFake.NewSimpleClientset()
	hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	gdc := &defaultGameDeploymentControl{hookRunLister: hookInformer.Tkex().V1alpha1().HookRuns().Lister()}
	deploy := test.NewGameDeployment(1)
	deploy.UID = "test"

	ht1 := test.NewHookTemplate()
	ht1.Name = "ht1"
	ht2 := test.NewHookTemplate()
	ht2.Name = "ht2"
	hr1 := test.NewHookRunFromTemplate(ht1, deploy)
	hr2 := test.NewHookRunFromTemplate(ht2, deploy)
	_ = hookInformer.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr1)
	_ = hookInformer.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr2)

	hrs, _ := gdc.getHookRunsForGameDeployment(deploy)
	if len(hrs) != 2 {
		t.Errorf("getHookRunsForGameDeployment should return 2 hookruns, but got %d", len(hrs))
	}
}

func expectPatchHookRunAction(namespace, name string, patch []byte) testing2.PatchActionImpl {
	return testing2.NewPatchAction(schema.GroupVersionResource{Group: "tkex.tencent.com", Version: v1alpha1.Version,
		Resource: "hookruns"}, namespace, name, types.MergePatchType, patch)
}

func expectDeleteHookRunAction(namespace, name string) testing2.DeleteActionImpl {
	return testing2.NewDeleteAction(schema.GroupVersionResource{Group: "tkex.tencent.com", Version: v1alpha1.Version,
		Resource: "hookruns"}, namespace, name)
}

func TestReconcileHookRuns(t *testing.T) {
	tests := []struct {
		name           string
		canaryCtx      *canaryContext
		expectedError  error
		expectedAction []testing2.Action
	}{
		{
			name: "cancel no step hookrun",
			canaryCtx: &canaryContext{
				deploy: func() *v1alpha1.GameDeployment {
					deploy := test.NewGameDeployment(1)
					deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{}
					deploy.Status.Canary.Revision = "1"
					deploy.Status.Canary.CurrentStepHookRun = "hr1"
					return deploy
				}(),
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdateRevision: "2",
				},
				currentHrs: []*alpha1.HookRun{
					newHR("hr1", v1alpha12.HookPhasePending, false, ""),
				},
			},
			expectedAction: []testing2.Action{
				expectPatchHookRunAction("default", "hr1", nil),
			},
		},
		{
			name: "create new hook run, but hook template is not found",
			canaryCtx: &canaryContext{
				deploy: func() *v1alpha1.GameDeployment {
					deploy := test.NewGameDeployment(1)
					deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
						Steps: []v1alpha1.CanaryStep{
							{Hook: &v1alpha12.HookStep{TemplateName: "foo", Args: []v1alpha12.HookRunArgument{
								{
									Name:  "foo",
									Value: "bar",
								},
							}}},
						},
					}
					deploy.Status.Canary.Revision = "1"
					deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
					return deploy
				}(),
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdateRevision: "2",
				},
			},
			expectedError: k8serrors.NewNotFound(v1alpha12.Resource("hooktemplate"), "foo"),
		},
		{
			name: "create new hook run with current hook run",
			canaryCtx: &canaryContext{
				deploy: func() *v1alpha1.GameDeployment {
					deploy := test.NewGameDeployment(1)
					deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
						Steps: []v1alpha1.CanaryStep{
							{Hook: &v1alpha12.HookStep{TemplateName: "foo"}},
						},
					}
					deploy.Status.Canary.Revision = "1"
					deploy.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
					deploy.Status.Canary.CurrentStepHookRun = "hr1"
					return deploy
				}(),
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdateRevision: "2",
				},
				currentHrs: []*alpha1.HookRun{
					newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
				},
			},
		},
		{
			name: "cancel hook run",
			canaryCtx: &canaryContext{
				deploy: func() *v1alpha1.GameDeployment {
					deploy := test.NewGameDeployment(1)
					deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{}
					deploy.Status.Canary.Revision = "1"
					deploy.Status.Canary.CurrentStepHookRun = "hr1"
					deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
						{Reason: "1"},
					}
					return deploy
				}(),
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdateRevision: "2",
				},
				otherHrs: []*v1alpha12.HookRun{
					newHR("hr2", v1alpha12.HookPhaseFailed, false, ""),
				},
				currentHrs: []*v1alpha12.HookRun{
					newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
				},
			},
			expectedAction: []testing2.Action{
				expectDeleteHookRunAction("default", "hr2"),
			},
		},
		{
			name: "delete hook run",
			canaryCtx: &canaryContext{
				deploy: func() *v1alpha1.GameDeployment {
					deploy := test.NewGameDeployment(1)
					deploy.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{}
					deploy.Status.Canary.Revision = "1"
					deploy.Status.Canary.CurrentStepHookRun = "hr2"
					deploy.Status.PauseConditions = []v1alpha12.PauseCondition{
						{Reason: "1"},
					}
					return deploy
				}(),
				newStatus: &v1alpha1.GameDeploymentStatus{
					UpdateRevision: "2",
				},
				otherHrs: []*v1alpha12.HookRun{
					newHR("hr2", v1alpha12.HookPhaseFailed, false, ""),
				},
				currentHrs: []*v1alpha12.HookRun{
					newHR("hr1", v1alpha12.HookPhaseFailed, false, ""),
				},
			},
			expectedAction: []testing2.Action{
				expectDeleteHookRunAction("default", "hr2"),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			hookClient := hookFake.NewSimpleClientset()
			hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
			gdc := &defaultGameDeploymentControl{
				hookRunLister:      hookInformer.Tkex().V1alpha1().HookRuns().Lister(),
				hookTemplateLister: hookInformer.Tkex().V1alpha1().HookTemplates().Lister(),
				hookClient:         hookClient,
			}

			err := gdc.reconcileHookRuns(s.canaryCtx)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("reconcileHookRuns should return: %v, but got: %v", s.expectedError, err)
			}
			if !test.EqualActions(s.expectedAction, test.FilterActions(hookClient.Actions(), test.FilterPatchAction)) {
				t.Errorf("expected actions: %v, but got: %v", s.expectedAction, hookClient.Actions())
			}
		})
	}
}

func TestNewHookRunFromGameDeployment(t *testing.T) {
	hookClient := hookFake.NewSimpleClientset()
	hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	gdc := &defaultGameDeploymentControl{
		hookRunLister:      hookInformer.Tkex().V1alpha1().HookRuns().Lister(),
		hookTemplateLister: hookInformer.Tkex().V1alpha1().HookTemplates().Lister(),
		hookClient:         hookClient,
	}

	deploy := test.NewGameDeployment(1)
	canaryCtx := &canaryContext{deploy: deploy}
	hookStep := &v1alpha12.HookStep{TemplateName: "hr"}
	revision := "1"
	stepIndex := int32(1)

	_, err := gdc.newHookRunFromGameDeployment(canaryCtx, hookStep, nil, revision, &stepIndex, nil)
	if !reflect.DeepEqual(err, k8serrors.NewNotFound(v1alpha12.Resource("hooktemplate"), "hr")) {
		t.Errorf("got error: %v", err)
	}
	template := test.NewHookTemplate()
	template.Name = "hr"
	_ = hookInformer.Tkex().V1alpha1().HookTemplates().Informer().GetIndexer().Add(template)
	hr, err := gdc.newHookRunFromGameDeployment(canaryCtx, hookStep, nil, revision, &stepIndex, nil)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if hr.Name != "canary-1-1-hr" {
		t.Errorf("name error, got: %s", hr.Name)
	}
	if hr.Labels != nil {
		t.Errorf("labels error, got: %v", hr.Labels)
	}
	if hr.Spec.Args != nil {
		t.Errorf("args error, got: %v", hr.Spec.Args)
	}
}
