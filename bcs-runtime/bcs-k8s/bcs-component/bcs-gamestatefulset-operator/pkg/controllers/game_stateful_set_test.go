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
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	stsfake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/clientset/versioned/fake"
	stsscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/clientset/versioned/scheme"
	stsinformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/requeueduration"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	"reflect"
	"testing"
	"time"
)

var alwaysReady = func() bool { return true }

type fixture struct {
	t testing.TB

	c *GameStatefulSetController

	kubeClient *fake.Clientset
	stsClient  *stsfake.Clientset
	hookClient *hookFake.Clientset
	// Objects to put in the store.
	stsLister        []*gstsv1alpha1.GameStatefulSet
	podLister        []*corev1.Pod
	controllerLister []*apps.ControllerRevision
	hookRunLister    []*hookv1alpha1.HookRun

	// informers
	kubeInformer informers.SharedInformerFactory
	stsInformer  stsinformers.SharedInformerFactory
	hookInformer hookInformers.SharedInformerFactory

	// Actions expected to happen on the client.
	kubeActions []core.Action
	stsActions  []core.Action
	hookActions []core.Action

	// Objects from here are also preloaded into NewSimpleFake.
	kubeObjects []runtime.Object
	stsObjects  []runtime.Object
	hookObjects []runtime.Object
}

func (f *fixture) expectGetPodAction(namespace, name string) {
	action := core.NewGetAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name)
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectUpdatePodAction(namespace string, object runtime.Object) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, object)
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectCreatePodAction(pod *corev1.Pod) {
	action := core.NewCreateAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		pod.Namespace, pod)
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectListPodAction(namespace string, opts metav1.ListOptions) {
	action := core.NewListAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		schema.GroupVersionKind{}, namespace, opts)
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectDeletePodAction(namespace, name string) {
	action := core.NewDeleteAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name)
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectPatchPodAction(namespace, name string) {
	action := core.NewPatchAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name, types.MergePatchType, []byte{})
	f.kubeActions = append(f.kubeActions, action)
}

func (f *fixture) expectGetGameStatefulSetAction(namespace, name string) {
	action := core.NewGetAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, namespace, name)
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectUpdateGameStatefulSetStatusAction(sts *gstsv1alpha1.GameStatefulSet) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, sts.Namespace, sts)
	action.Subresource = "status"
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectListGameStatefulSetActions(namespace string, opts metav1.ListOptions) {
	action := core.NewListAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, schema.GroupVersionKind{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Kind: gstsv1alpha1.Kind}, namespace, opts)
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectWatchGameStatefulSetActions(namespace string) {
	action := core.NewWatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, namespace, metav1.ListOptions{})
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectPatchGameStatefulSetAction(sts *gstsv1alpha1.GameStatefulSet, patch []byte) {
	action := core.NewPatchAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, sts.Namespace, sts.Name, types.MergePatchType, patch)
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectPatchGameStatefulSetSubResourceAction(sts *gstsv1alpha1.GameStatefulSet, patch []byte) {
	action := core.NewPatchSubresourceAction(schema.GroupVersionResource{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version,
		Resource: gstsv1alpha1.Plural}, sts.Namespace, sts.Name, types.MergePatchType, patch, "status")
	f.stsActions = append(f.stsActions, action)
}

func (f *fixture) expectCreateControllerRevision(cr *apps.ControllerRevision) {
	action := core.NewCreateAction(schema.GroupVersionResource{Group: apps.GroupName, Version: apps.SchemeGroupVersion.Version,
		Resource: "controllerrevisions"}, cr.Namespace, cr)
	f.kubeActions = append(f.kubeActions, action)
}

func newFixture(t testing.TB) *fixture {
	_ = stsscheme.AddToScheme(scheme.Scheme)
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	f.stsObjects = []runtime.Object{}
	f.hookObjects = []runtime.Object{}
	return f
}

func (f *fixture) newController() {
	// reset expectations cache
	scaleExpectations = expectations.NewScaleExpectations()
	durationStore = requeueduration.DurationStore{}

	// Create the controller
	f.kubeClient = fake.NewSimpleClientset(f.kubeObjects...)
	f.stsClient = stsfake.NewSimpleClientset(f.stsObjects...)
	f.hookClient = hookFake.NewSimpleClientset(f.hookObjects...)
	f.kubeInformer = informers.NewSharedInformerFactory(f.kubeClient, controller.NoResyncPeriodFunc())
	f.stsInformer = stsinformers.NewSharedInformerFactory(f.stsClient, controller.NoResyncPeriodFunc())
	f.hookInformer = hookInformers.NewSharedInformerFactory(f.hookClient, controller.NoResyncPeriodFunc())

	c := NewGameStatefulSetController(
		f.kubeInformer.Core().V1().Pods(),
		f.stsInformer.Tkex().V1alpha1().GameStatefulSets(),
		f.kubeInformer.Core().V1().PersistentVolumeClaims(),
		f.kubeInformer.Core().V1().Nodes(),
		f.kubeInformer.Apps().V1().ControllerRevisions(),
		f.hookInformer.Tkex().V1alpha1().HookRuns(),
		f.hookInformer.Tkex().V1alpha1().HookTemplates(),
		f.kubeClient, f.stsClient, f.hookClient)
	c.podListerSynced = alwaysReady
	c.setListerSynced = alwaysReady
	c.revListerSynced = alwaysReady
	for _, pod := range f.podLister {
		_ = f.kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
		_, _ = f.kubeClient.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	}
	for _, r := range f.controllerLister {
		_ = f.kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().Add(r)
		_, _ = f.kubeClient.AppsV1().ControllerRevisions(r.Namespace).Create(context.TODO(), r, metav1.CreateOptions{})
	}
	for _, d := range f.stsLister {
		_ = f.stsInformer.Tkex().V1alpha1().GameStatefulSets().Informer().GetIndexer().Add(d)
		_, _ = f.stsClient.TkexV1alpha1().GameStatefulSets(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
	}

	// remove init test data
	f.kubeClient.ClearActions()
	f.stsClient.ClearActions()
	f.c = c
}

func (f *fixture) runExpectError(objectName string, startInformers bool) {
	f.run_(objectName, startInformers, true)
}

func (f *fixture) run(objectName string) {
	f.run_(objectName, true, false)
}

func (f *fixture) run_(objectName string, startInformers bool, expectError bool) {
	f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		f.stsInformer.Start(stopCh)
		f.kubeInformer.Start(stopCh)
	}

	err := f.c.sync(objectName)
	if !expectError && err != nil {
		if err.Error() != "PatchType is not supported" {
			f.t.Errorf("error syncing gamestatefulset: %v", err)
		}
	} else if expectError && err == nil {
		f.t.Error("expected error syncing gamestatefulset, got nil")
	}

	assertActions(f.kubeActions, filterInformerActions(f.kubeClient.Actions()), f.t)
	assertActions(f.stsActions, filterInformerActions(f.stsClient.Actions()), f.t)
}

func assertActions(expect, got []core.Action, t testing.TB) {
	for i, action := range got {
		if len(expect) < i+1 {
			t.Errorf("%d unexpected actions: %+v", len(got)-len(expect), got[i:])
			break
		}

		expectedAction := expect[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) && action.GetSubresource() == expectedAction.GetSubresource()) {
			t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(expect) > len(got) {
		t.Errorf("%d additional expected actions:%+v", len(expect)-len(got), expect[len(got):])
	}
}

func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "pods") ||
				action.Matches("list", "deployments") ||
				action.Matches("list", "replicasets") ||
				action.Matches("list", "gamestatefulsets") ||
				action.Matches("list", "controllerrevisions") ||
				action.Matches("list", "persistentvolumeclaims") ||
				action.Matches("list", "nodes") ||
				action.Matches("watch", "pods") ||
				action.Matches("watch", "gamestatefulsets") ||
				action.Matches("watch", "deployments") ||
				action.Matches("watch", "controllerrevisions") ||
				action.Matches("watch", "persistentvolumeclaims") ||
				action.Matches("watch", "nodes") ||
				action.Matches("watch", "replicasets") ||
				action.Matches("create", "events")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func newPod(suffix interface{}, labels map[string]string, running bool) *corev1.Pod {
	pod := testutil.NewPod(suffix)
	pod.Labels = labels
	if running {
		pod.Status.Phase = corev1.PodRunning
	}
	return pod
}

func TestSyncGameStatefulSetCreatePods(t *testing.T) {
	f := newFixture(t)
	sts := testutil.NewGameStatefulSet(3)
	f.stsObjects = append(f.stsObjects, sts)

	// create a simple game statefulSet, it should create 3 pods and 1 controller revision
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 1))
	// create one pod and wait pod to running
	f.expectCreatePodAction(testutil.NewPod(0))
	// update game statefulset status
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)

	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

// Test RollingUpdate, rolling update will delete pods to update game deployment.
func TestSyncRollingUpdate(t *testing.T) {
	f := newFixture(t)
	sts := testutil.NewGameStatefulSet(1)
	sts.Spec.UpdateStrategy = gstsv1alpha1.GameStatefulSetUpdateStrategy{
		Type: gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType,
	}
	// pod and controller revision is existed.
	f.stsObjects = append(f.stsObjects, sts)
	pod := newPod(0, sts.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)
	cr := testutil.NewControllerRevision(sts, 1)
	f.controllerLister = append(f.controllerLister, cr)

	// for the first sync, it will get object with kube client
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// get the new object for claiming pods
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 2))
	f.expectUpdatePodAction(pod.Namespace, pod)
	// delete pod to update
	f.expectDeletePodAction(pod.Namespace, pod.Name)
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)
	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

// Test InPlaceUpdate, inplace update will not delete pods to update game statefulset.
func TestSyncInPlaceUpdate(t *testing.T) {
	f := newFixture(t)
	sts := testutil.NewGameStatefulSet(1)
	sts.Spec.UpdateStrategy = gstsv1alpha1.GameStatefulSetUpdateStrategy{
		Type: gstsv1alpha1.InplaceUpdateGameStatefulSetStrategyType,
	}
	// pod and controller revision is existed.
	f.stsObjects = append(f.stsObjects, sts)
	cr, err := newRevision(sts, 1, func() *int32 { a := int32(1); return &a }())
	if err != nil {
		t.Fatalf("create revision error: %v", err)
	}
	labels := sts.Spec.Template.Labels
	labels[apps.ControllerRevisionHashLabelKey] = cr.Name
	pod := newPod(0, sts.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)
	f.controllerLister = append(f.controllerLister, cr)

	// for the first sync, it will get object with kube client
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// get the new object for claiming pods
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 2))
	//f.expectGetPodAction(pod.Namespace, pod.Name)
	f.expectUpdatePodAction(pod.Namespace, nil)
	f.expectGetPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)
	f.expectUpdatePodAction(pod.Namespace, nil)
	f.expectGetPodAction(pod.Namespace, pod.Name)
	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

// Test canary step
func TestSyncCanaryStep(t *testing.T) {
	f := newFixture(t)

	sts := testutil.NewGameStatefulSet(1)
	sts.Spec.UpdateStrategy = gstsv1alpha1.GameStatefulSetUpdateStrategy{
		Type: gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType,
		CanaryStrategy: &gstsv1alpha1.CanaryStrategy{
			Steps: []gstsv1alpha1.CanaryStep{
				{Pause: &gstsv1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
			},
		},
	}
	// pod and controller revision is existed.
	f.stsObjects = append(f.stsObjects, sts)

	pod := newPod(0, sts.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// get the new object for claiming pods
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 2))
	f.expectUpdatePodAction(pod.Namespace, pod)
	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

func TestSyncCanaryStepChange(t *testing.T) {
	f := newFixture(t)

	sts := testutil.NewGameStatefulSet(1)
	sts.Spec.UpdateStrategy = gstsv1alpha1.GameStatefulSetUpdateStrategy{
		Type: gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType,
		CanaryStrategy: &gstsv1alpha1.CanaryStrategy{
			Steps: []gstsv1alpha1.CanaryStep{
				{Pause: &gstsv1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
				{Pause: &gstsv1alpha1.CanaryPause{Duration: func() *int32 { a := int32(2); return &a }()}},
			},
		},
	}
	sts.Status.CurrentStepHash = "123"
	sts.Status.UpdateRevision = "2"

	// pod and controller revision is existed.
	f.stsObjects = append(f.stsObjects, sts)

	pod := newPod(0, sts.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// get the new object for claiming pods
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 2))
	// update game statefulSet status
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)
	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

func TestSyncRevisionChange(t *testing.T) {
	f := newFixture(t)

	sts := testutil.NewGameStatefulSet(1)
	sts.Spec.UpdateStrategy = gstsv1alpha1.GameStatefulSetUpdateStrategy{
		Type: gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType,
		CanaryStrategy: &gstsv1alpha1.CanaryStrategy{
			Steps: []gstsv1alpha1.CanaryStep{
				{Pause: &gstsv1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
				{Pause: &gstsv1alpha1.CanaryPause{Duration: func() *int32 { a := int32(2); return &a }()}},
			},
		},
	}
	sts.Status.UpdateRevision = "2"

	// pod and controller revision is existed.
	f.stsObjects = append(f.stsObjects, sts)

	pod := newPod(0, sts.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// get the new object for claiming pods
	f.expectGetGameStatefulSetAction(sts.Namespace, sts.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	f.expectCreateControllerRevision(testutil.NewControllerRevision(sts, 2))
	// update game deployment status
	f.expectPatchGameStatefulSetSubResourceAction(sts, nil)
	f.run(types.NamespacedName{
		Namespace: sts.Namespace,
		Name:      sts.Name,
	}.String())
}

// Test delete pod
func TestDeletePod(t *testing.T) {
	f := newFixture(t)
	sts := testutil.NewGameStatefulSet(3)
	f.stsLister = append(f.stsLister, sts)
	pod := newPod(0, sts.Spec.Template.Labels, true)
	pod.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	f.podLister = append(f.podLister, pod)

	sts2 := testutil.NewGameStatefulSet(3)
	sts2.Name = "test-2"
	pod3 := newPod(1, sts2.Spec.Template.Labels, true)
	pod3.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts2, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	f.podLister = append(f.podLister, pod3)

	f.newController()
	f.c.deletePod(pod)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := f.c.queue.Get()
	if key == nil || done {
		t.Fatalf("failed to enqueue controller for pod %v", pod.Name)
	}
	expectedKey, _ := controller.KeyFunc(sts)
	if got, want := key.(string), expectedKey; got != want {
		t.Errorf("queue.Get() = %v, want %v", got, want)
	}

	// pod with not owner reference
	pod2 := testutil.NewPod(1)
	f.c.deletePod(pod2)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// not pod
	f.c.deletePod(sts)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// pod with owner reference but sts not exist
	f.c.deletePod(pod3)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// deleted pod
	a := cache.DeletedFinalStateUnknown{}
	f.c.deletePod(a)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestEnqueueGameStatefulSetForHook(t *testing.T) {
	f := newFixture(t)
	sts := testutil.NewGameStatefulSet(3)
	hr := testutil.NewHookRunFromTemplate(testutil.NewHookTemplate(), sts)
	f.stsLister = append(f.stsLister, sts)

	f.newController()

	f.c.enqueueGameStatefulSetForHook(hr)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	// remove the key from the queue
	for f.c.queue.Len() > 0 {
		f.c.queue.Get()
	}

	a := cache.DeletedFinalStateUnknown{}
	f.c.enqueueGameStatefulSetForHook(a)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestAddPod(t *testing.T) {
	f := newFixture(t)

	sts := testutil.NewGameStatefulSet(3)
	f.stsLister = append(f.stsLister, sts)
	pod := newPod(0, sts.Spec.Template.Labels, true)
	pod.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	pod2 := newPod(1, sts.Spec.Template.Labels, true)

	f.newController()
	f.c.addPod(pod)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := f.c.queue.Get()
	if key == nil || done {
		t.Fatalf("failed to enqueue controller for pod %v", pod.Name)
	}
	expectedKey, _ := controller.KeyFunc(sts)
	if got, want := key.(string), expectedKey; got != want {
		t.Errorf("queue.Get() = %v, want %v", got, want)
	}

	// not controller reference
	f.c.addPod(pod2)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// already deleted
	pod2.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	f.c.addPod(pod2)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestUpdatePod(t *testing.T) {
	f := newFixture(t)

	sts := testutil.NewGameStatefulSet(3)
	sts2 := testutil.NewGameStatefulSet(3)
	sts2.Name = "test-2"
	sts3 := testutil.NewGameStatefulSet(3)
	sts3.Name = "test-3"
	sts4 := testutil.NewGameStatefulSet(3)
	sts4.Name = "test-4"
	f.stsLister = append(f.stsLister, sts)
	f.stsLister = append(f.stsLister, sts2)
	f.stsLister = append(f.stsLister, sts3)
	f.stsLister = append(f.stsLister, sts4)
	cur := newPod(0, sts.Spec.Template.Labels, true)
	cur.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	old := newPod(1, sts.Spec.Template.Labels, true)
	cur2 := cur.DeepCopy()
	cur2.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts3, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	cur2.ResourceVersion = "cur2"
	cur2.Name = "cur2"
	cur3 := cur.DeepCopy()
	cur3.OwnerReferences = nil
	cur3.Labels = sts4.Spec.Template.Labels
	cur3.ResourceVersion = "cur3"
	cur3.Name = "cur3"
	cur3.Labels["cur3"] = "true"

	f.newController()

	// same pod resource version, won't enqueue controller
	f.c.updatePod(old, cur)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// cur pod has been deleted
	cur.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	cur.Labels = map[string]string{"foo": "cur"}
	cur.ResourceVersion = "cur"
	f.c.updatePod(old, cur)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	for f.c.queue.Len() > 0 {
		f.c.queue.Get()
	}

	// controller reference is changed, will enqueue controller
	old.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts2, gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet"))}
	f.c.updatePod(old, cur2)
	if got, want := f.c.queue.Len(), 2; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	for f.c.queue.Len() > 0 {
		f.c.queue.Get()
	}

	// label or owner reference is changed, will enqueue controller
	f.c.updatePod(old, cur3)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestSetAdoptOrphanRevisions(t *testing.T) {
	stsfake.AddToScheme(scheme.Scheme)
	tests := []struct {
		name            string
		set             *gstsv1alpha1.GameStatefulSet
		existSet        []*gstsv1alpha1.GameStatefulSet
		existRevisions  []*apps.ControllerRevision
		expectedError   error
		expectedActions []core.Action
	}{
		{
			name: "adopt with no gamestatefulsets",
			set:  testutil.NewGameStatefulSet(2),
			existRevisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					cr := newControllerRevision("foo-0")
					cr.Labels = map[string]string{"foo": "bar"}
					return cr
				}(),
				func() *apps.ControllerRevision {
					cr := newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(2))
					cr.OwnerReferences = nil
					return cr
				}(),
				newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(2)),
			},
			expectedError: k8serrors.NewNotFound(gstsv1alpha1.SchemeGroupVersion.WithResource("gamestatefulsets").GroupResource(), "foo"),
		},
		{
			name: "adopt with wrong uid",
			set:  testutil.NewGameStatefulSet(2),
			existSet: []*gstsv1alpha1.GameStatefulSet{
				func() *gstsv1alpha1.GameStatefulSet {
					set := testutil.NewGameStatefulSet(2)
					set.UID = "2"
					return set
				}(),
			},
			existRevisions: []*apps.ControllerRevision{
				func() *apps.ControllerRevision {
					cr := newControllerRevision("foo-0")
					cr.Labels = map[string]string{"foo": "bar"}
					return cr
				}(),
				func() *apps.ControllerRevision {
					cr := newCRWithName("foo-1", 1, testutil.NewGameStatefulSet(2))
					cr.OwnerReferences = nil
					return cr
				}(),
				newCRWithName("foo-2", 1, testutil.NewGameStatefulSet(2)),
			},
			expectedError: errors.New("original GameStatefulSet default/foo is gone: got uid 2, wanted test"),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			gstsClient := stsfake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			informer := informerFactory.Apps().V1().ControllerRevisions()
			for i := range s.existRevisions {
				informer.Informer().GetIndexer().Add(s.existRevisions[i])
			}
			for i := range s.existSet {
				gstsClient.TkexV1alpha1().GameStatefulSets(s.existSet[i].Namespace).Create(context.TODO(), s.existSet[i], metav1.CreateOptions{})
			}
			gstsClient.ClearActions()
			controllerHistory := history.NewFakeHistory(informer)
			control := &defaultGameStatefulSetControl{
				controllerHistory: controllerHistory,
			}
			ssc := &GameStatefulSetController{
				gstsClient: gstsClient,
				control:    control,
			}
			err := ssc.adoptOrphanRevisions(s.set)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("err: %v, expected: %v", err, s.expectedError)
			}
		})
	}
}
