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
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"testing"
	"time"
)

func TestGetHookRunFromGameStatefulSet(t *testing.T) {
	hookClient := hookFake.NewSimpleClientset()
	hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	ssc := &defaultGameStatefulSetControl{hookRunLister: hookInformer.Tkex().V1alpha1().HookRuns().Lister()}
	set := testutil.NewGameStatefulSet(1)

	ht1 := testutil.NewHookTemplate()
	ht1.Name = "ht1"
	ht2 := testutil.NewHookTemplate()
	ht2.Name = "ht2"
	hr1 := testutil.NewHookRunFromTemplate(ht1, set)
	hr2 := testutil.NewHookRunFromTemplate(ht2, set)
	hookInformer.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr1)
	hookInformer.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(hr2)

	hrs, _ := ssc.getHookRunsForGameStatefulSet(set)
	if len(hrs) != 2 {
		t.Errorf("should return 2 hookruns, but got %d", len(hrs))
	}
}

func expectPatchHookRunAction(namespace, name string, patch []byte) testing2.PatchActionImpl {
	return testing2.NewPatchAction(schema.GroupVersionResource{Group: "tkex.tencent.com", Version: "v1alpha1",
		Resource: "hookruns"}, namespace, name, types.MergePatchType, patch)
}

func expectDeleteHookRunAction(namespace, name string) testing2.DeleteActionImpl {
	return testing2.NewDeleteAction(schema.GroupVersionResource{Group: "tkex.tencent.com", Version: "v1alpha1",
		Resource: "hookruns"}, namespace, name)
}

func newHR(name string, phase hookv1alpha1.HookPhase, deleted bool, hrType string) *hookv1alpha1.HookRun {
	hr := testutil.NewHookRunFromTemplate(testutil.NewHookTemplate(), testutil.NewGameStatefulSet(1))
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
				set: func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(1)
					set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{}
					set.Status.Canary.Revision = "1"
					set.Status.Canary.CurrentStepHookRun = "hr1"
					return set
				}(),
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdateRevision: "2",
				},
				currentHrs: []*hookv1alpha1.HookRun{
					newHR("hr1", hookv1alpha1.HookPhasePending, false, ""),
				},
			},
			expectedAction: []testing2.Action{
				expectPatchHookRunAction("default", "hr1", nil),
			},
		},
		{
			name: "create new hook run, but hook template is not found",
			canaryCtx: &canaryContext{
				set: func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(1)
					set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
						Steps: []gstsv1alpha1.CanaryStep{
							{Hook: &hookv1alpha1.HookStep{TemplateName: "foo", Args: []hookv1alpha1.HookRunArgument{
								{
									Name:  "foo",
									Value: "bar",
								},
							}}},
						},
					}
					set.Status.Canary.Revision = "1"
					set.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
					return set
				}(),
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdateRevision: "2",
				},
			},
			expectedError: k8serrors.NewNotFound(hookv1alpha1.Resource("hooktemplate"), "foo"),
		},
		{
			name: "create new hook run with current hook run",
			canaryCtx: &canaryContext{
				set: func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(1)
					set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
						Steps: []gstsv1alpha1.CanaryStep{
							{Hook: &hookv1alpha1.HookStep{TemplateName: "foo"}},
						},
					}
					set.Status.Canary.Revision = "1"
					set.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
					set.Status.Canary.CurrentStepHookRun = "hr1"
					return set
				}(),
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdateRevision: "2",
				},
				currentHrs: []*hookv1alpha1.HookRun{
					newHR("hr1", hookv1alpha1.HookPhaseFailed, false, ""),
				},
			},
		},
		{
			name: "cancel hook run",
			canaryCtx: &canaryContext{
				set: func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(1)
					set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{}
					set.Status.Canary.Revision = "1"
					set.Status.Canary.CurrentStepHookRun = "hr1"
					set.Status.PauseConditions = []hookv1alpha1.PauseCondition{
						{Reason: "1"},
					}
					return set
				}(),
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdateRevision: "2",
				},
				otherHrs: []*hookv1alpha1.HookRun{
					newHR("hr2", hookv1alpha1.HookPhaseFailed, false, ""),
				},
				currentHrs: []*hookv1alpha1.HookRun{
					newHR("hr1", hookv1alpha1.HookPhaseFailed, false, ""),
				},
			},
			expectedAction: []testing2.Action{
				expectDeleteHookRunAction("default", "hr2"),
			},
		},
		{
			name: "delete hook run",
			canaryCtx: &canaryContext{
				set: func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(1)
					set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{}
					set.Status.Canary.Revision = "1"
					set.Status.Canary.CurrentStepHookRun = "hr2"
					set.Status.PauseConditions = []hookv1alpha1.PauseCondition{
						{Reason: "1"},
					}
					return set
				}(),
				newStatus: &gstsv1alpha1.GameStatefulSetStatus{
					UpdateRevision: "2",
				},
				otherHrs: []*hookv1alpha1.HookRun{
					newHR("hr2", hookv1alpha1.HookPhaseFailed, false, ""),
				},
				currentHrs: []*hookv1alpha1.HookRun{
					newHR("hr1", hookv1alpha1.HookPhaseFailed, false, ""),
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
			ssc := &defaultGameStatefulSetControl{
				hookRunLister:      hookInformer.Tkex().V1alpha1().HookRuns().Lister(),
				hookTemplateLister: hookInformer.Tkex().V1alpha1().HookTemplates().Lister(),
				hookClient:         hookClient,
			}

			err := ssc.reconcileHookRuns(s.canaryCtx)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("reconcileHookRuns should return: %v, but got: %v", s.expectedError, err)
			}
			if !testutil.EqualActions(s.expectedAction, testutil.FilterActions(hookClient.Actions(), testutil.FilterPatchAction)) {
				t.Errorf("expected actions: %v, but got: %v", s.expectedAction, hookClient.Actions())
			}
		})
	}
}

func TestNewHookRunFromGameStatefulSet(t *testing.T) {
	hookClient := hookFake.NewSimpleClientset()
	hookInformer := hookInformers.NewSharedInformerFactory(hookClient, controller.NoResyncPeriodFunc())
	ssc := &defaultGameStatefulSetControl{
		hookRunLister:      hookInformer.Tkex().V1alpha1().HookRuns().Lister(),
		hookTemplateLister: hookInformer.Tkex().V1alpha1().HookTemplates().Lister(),
		hookClient:         hookClient,
	}

	set := testutil.NewGameStatefulSet(1)
	canaryCtx := &canaryContext{set: set}
	hookStep := &hookv1alpha1.HookStep{TemplateName: "hr"}
	revision := "1"
	stepIndex := int32(1)

	_, err := ssc.newHookRunFromGameStatefulSet(canaryCtx, hookStep, nil, revision, &stepIndex, nil)
	if !reflect.DeepEqual(err, k8serrors.NewNotFound(hookv1alpha1.Resource("hooktemplate"), "hr")) {
		t.Errorf("got error: %v", err)
	}
	template := testutil.NewHookTemplate()
	template.Name = "hr"
	hookInformer.Tkex().V1alpha1().HookTemplates().Informer().GetIndexer().Add(template)
	hr, err := ssc.newHookRunFromGameStatefulSet(canaryCtx, hookStep, nil, revision, &stepIndex, nil)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if hr.Name != "canary-1-step1-hr" {
		t.Errorf("name error, got: %s", hr.Name)
	}
	if hr.Labels != nil {
		t.Errorf("labels error, got: %v", hr.Labels)
	}
	if hr.Spec.Args != nil {
		t.Errorf("args error, got: %v", hr.Spec.Args)
	}
}
