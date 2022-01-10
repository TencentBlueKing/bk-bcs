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
	"context"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	deployFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/fake"
	gdscheme "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned/scheme"
	deployInformer "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/revision"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/expectations"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/controller/history"
	"testing"
	"time"
)

var alwaysReady = func() bool { return true }

type fixture struct {
	t testing.TB

	c *GameDeploymentController

	kubeClient   *fake.Clientset
	deployClient *deployFake.Clientset
	hookClient   *hookFake.Clientset
	// Objects to put in the store.
	deployLister     []*v1alpha1.GameDeployment
	podLister        []*corev1.Pod
	controllerLister []*apps.ControllerRevision
	hookRunLister    []*hookv1alpha1.HookRun

	// informers
	kubeInformer   informers.SharedInformerFactory
	deployInformer deployInformer.SharedInformerFactory
	hookInformer   hookInformers.SharedInformerFactory

	// Actions expected to happen on the client.
	kubeActions   []core.Action
	deployActions []core.Action
	hookActions   []core.Action

	// Objects from here are also preloaded into NewSimpleFake.
	kubeObjects   []runtime.Object
	deployObjects []runtime.Object
	hookObjects   []runtime.Object
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

func (f *fixture) expectGetGameDeploymentAction(namespace, name string) {
	action := core.NewGetAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, namespace, name)
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectUpdateGameDeploymentStatusAction(d *v1alpha1.GameDeployment) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, d.Namespace, d)
	action.Subresource = "status"
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectUpdateGameDeploymentAction(d *apps.Deployment) {
	action := core.NewUpdateAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, d.Namespace, d)
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectListGameDeploymentActions(namespace string, opts metav1.ListOptions) {
	action := core.NewListAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, schema.GroupVersionKind{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Kind: v1alpha1.Kind}, namespace, opts)
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectWatchGameDeploymentActions(namespace string) {
	action := core.NewWatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, namespace, metav1.ListOptions{})
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectPatchGameDeploymentAction(deploy *v1alpha1.GameDeployment, patch []byte) {
	action := core.NewPatchAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, deploy.Namespace, deploy.Name, types.MergePatchType, patch)
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectPatchGameDeploymentSubResourceAction(deploy *v1alpha1.GameDeployment, patch []byte) {
	action := core.NewPatchSubresourceAction(schema.GroupVersionResource{Group: v1alpha1.GroupName, Version: v1alpha1.Version,
		Resource: v1alpha1.Plural}, deploy.Namespace, deploy.Name, types.MergePatchType, patch, "status")
	f.deployActions = append(f.deployActions, action)
}

func (f *fixture) expectCreateControllerRevision(cr *apps.ControllerRevision) {
	action := core.NewCreateAction(schema.GroupVersionResource{Group: apps.GroupName, Version: apps.SchemeGroupVersion.Version,
		Resource: "controllerrevisions"}, cr.Namespace, cr)
	f.kubeActions = append(f.kubeActions, action)
}

func newFixture(t testing.TB) *fixture {
	_ = gdscheme.AddToScheme(scheme.Scheme)
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	f.deployObjects = []runtime.Object{}
	f.hookObjects = []runtime.Object{}
	return f
}

func (f *fixture) newController() {
	// reset expectations cache
	scaleExpectations = expectations.NewScaleExpectations()
	updateExpectations = expectations.NewUpdateExpectations(util.GetPodRevision)

	// Create the controller
	f.kubeClient = fake.NewSimpleClientset(f.kubeObjects...)
	f.deployClient = deployFake.NewSimpleClientset(f.deployObjects...)
	f.hookClient = hookFake.NewSimpleClientset(f.hookObjects...)
	f.kubeInformer = informers.NewSharedInformerFactory(f.kubeClient, controller.NoResyncPeriodFunc())
	f.deployInformer = deployInformer.NewSharedInformerFactory(f.deployClient, controller.NoResyncPeriodFunc())
	f.hookInformer = hookInformers.NewSharedInformerFactory(f.hookClient, controller.NoResyncPeriodFunc())

	c := NewGameDeploymentController(f.kubeInformer.Core().V1().Pods(), f.deployInformer.Tkex().V1alpha1().GameDeployments(),
		f.hookInformer.Tkex().V1alpha1().HookRuns(), f.hookInformer.Tkex().V1alpha1().HookTemplates(),
		f.kubeInformer.Apps().V1().ControllerRevisions(), f.kubeClient, f.deployClient, &record.FakeRecorder{}, f.hookClient,
		history.NewFakeHistory(f.kubeInformer.Apps().V1().ControllerRevisions()))
	c.podListerSynced = alwaysReady
	c.gdListerSynced = alwaysReady
	c.revListerSynced = alwaysReady
	for _, pod := range f.podLister {
		_ = f.kubeInformer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
		_, _ = f.kubeClient.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	}
	for _, r := range f.controllerLister {
		_ = f.kubeInformer.Apps().V1().ControllerRevisions().Informer().GetIndexer().Add(r)
		_, _ = f.kubeClient.AppsV1().ControllerRevisions(r.Namespace).Create(context.TODO(), r, metav1.CreateOptions{})
	}
	for _, d := range f.deployLister {
		_ = f.deployInformer.Tkex().V1alpha1().GameDeployments().Informer().GetIndexer().Add(d)
		_, _ = f.deployClient.TkexV1alpha1().GameDeployments(d.Namespace).Create(context.TODO(), d, metav1.CreateOptions{})
	}

	// remove init test data
	f.kubeClient.ClearActions()
	f.deployClient.ClearActions()
	f.c = c
}

func (f *fixture) runExpectError(deploymentName string, startInformers bool) {
	f.run_(deploymentName, startInformers, true)
}

func (f *fixture) run(deploymentName string) {
	f.run_(deploymentName, true, false)
}

func (f *fixture) run_(deploymentName string, startInformers bool, expectError bool) {
	f.newController()
	if startInformers {
		stopCh := make(chan struct{})
		defer close(stopCh)
		f.deployInformer.Start(stopCh)
		f.kubeInformer.Start(stopCh)
	}

	err := f.c.sync(deploymentName)
	if !expectError && err != nil {
		if err.Error() != "PatchType is not supported" {
			f.t.Errorf("error syncing gamedeployment: %v", err)
		}
	} else if expectError && err == nil {
		f.t.Error("expected error syncing gamedeployment, got nil")
	}

	assertActions(f.kubeActions, filterInformerActions(f.kubeClient.Actions()), f.t)
	assertActions(f.deployActions, filterInformerActions(f.deployClient.Actions()), f.t)
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
				action.Matches("list", "gamedeployments") ||
				action.Matches("list", "controllerrevisions") ||
				action.Matches("watch", "pods") ||
				action.Matches("watch", "gamedeployments") ||
				action.Matches("watch", "deployments") ||
				action.Matches("watch", "controllerrevisions") ||
				action.Matches("watch", "replicasets")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func newPod(suffix interface{}, labels map[string]string, running bool) *corev1.Pod {
	pod := test.NewPod(suffix)
	pod.Labels = labels
	if running {
		pod.Status.Phase = corev1.PodRunning
	}
	return pod
}

func TestSyncGameDeploymentCreatePods(t *testing.T) {
	f := newFixture(t)
	deploy := test.NewGameDeployment(3)
	f.deployObjects = append(f.deployObjects, deploy)

	// create a simple game deployment, it should create 3 pods and 1 controller revision
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// controller revision created by fake informer, so we don't need to expect it in the client actions
	//f.expectCreateControllerRevision(test.NewGDControllerRevision(deploy, 1))
	f.expectCreatePodAction(test.NewPod(0))
	f.expectCreatePodAction(test.NewPod(1))
	f.expectCreatePodAction(test.NewPod(2))
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)

	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

// Test RollingUpdate, rolling update will delete pods to update game deployment.
func TestSyncRollingUpdate(t *testing.T) {
	f := newFixture(t)
	deploy := test.NewGameDeployment(1)
	deploy.Spec.UpdateStrategy = v1alpha1.GameDeploymentUpdateStrategy{
		Type: v1alpha1.RollingGameDeploymentUpdateStrategyType,
	}
	// pod and controller revision is existed.
	f.deployObjects = append(f.deployObjects, deploy)
	pod := newPod(0, deploy.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)
	cr := test.NewGDControllerRevision(deploy, 1)
	f.controllerLister = append(f.controllerLister, cr)

	// for the first sync, it will get object with kube client
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// get the new object for claiming pods
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// delete pod to update
	f.expectDeletePodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)
	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

// Test InPlaceUpdate, inplace update will not delete pods to update game deployment.
func TestSyncInPlaceUpdate(t *testing.T) {
	f := newFixture(t)
	deploy := test.NewGameDeployment(1)
	deploy.Spec.UpdateStrategy = v1alpha1.GameDeploymentUpdateStrategy{
		Type: v1alpha1.InPlaceGameDeploymentUpdateStrategyType,
	}
	// pod and controller revision is existed.
	f.deployObjects = append(f.deployObjects, deploy)
	control := revision.NewRevisionControl()
	cr, err := control.NewRevision(deploy, 1, func() *int32 { a := int32(1); return &a }())
	if err != nil {
		t.Fatalf("create revision error: %v", err)
	}
	labels := deploy.Spec.Template.Labels
	labels[apps.ControllerRevisionHashLabelKey] = cr.Name
	pod := newPod(0, deploy.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)
	//cr := test.NewGDControllerRevision(deploy, 1)
	f.controllerLister = append(f.controllerLister, cr)

	// for the first sync, it will get object with kube client
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// get the new object for claiming pods
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	f.expectGetPodAction(pod.Namespace, pod.Name)
	f.expectUpdatePodAction(pod.Namespace, nil)
	f.expectGetPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)
	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

// Test canary step
func TestSyncCanaryStep(t *testing.T) {
	f := newFixture(t)

	deploy := test.NewGameDeployment(1)
	deploy.Spec.UpdateStrategy = v1alpha1.GameDeploymentUpdateStrategy{
		Type: v1alpha1.RollingGameDeploymentUpdateStrategyType,
		CanaryStrategy: &v1alpha1.CanaryStrategy{
			Steps: []v1alpha1.CanaryStep{
				{Pause: &v1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
			},
		},
	}
	// pod and controller revision is existed.
	f.deployObjects = append(f.deployObjects, deploy)

	pod := newPod(0, deploy.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// get the new object for claiming pods
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)
	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

func TestSyncCanaryStepChange(t *testing.T) {
	f := newFixture(t)

	deploy := test.NewGameDeployment(1)
	deploy.Spec.UpdateStrategy = v1alpha1.GameDeploymentUpdateStrategy{
		Type: v1alpha1.RollingGameDeploymentUpdateStrategyType,
		CanaryStrategy: &v1alpha1.CanaryStrategy{
			Steps: []v1alpha1.CanaryStep{
				{Pause: &v1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
				{Pause: &v1alpha1.CanaryPause{Duration: func() *int32 { a := int32(2); return &a }()}},
			},
		},
	}
	deploy.Status.CurrentStepHash = "123"
	deploy.Status.UpdateRevision = "2"

	// pod and controller revision is existed.
	f.deployObjects = append(f.deployObjects, deploy)

	pod := newPod(0, deploy.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// get the new object for claiming pods
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)
	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

func TestSyncRevisionChange(t *testing.T) {
	f := newFixture(t)

	deploy := test.NewGameDeployment(1)
	deploy.Spec.UpdateStrategy = v1alpha1.GameDeploymentUpdateStrategy{
		Type: v1alpha1.RollingGameDeploymentUpdateStrategyType,
		CanaryStrategy: &v1alpha1.CanaryStrategy{
			Steps: []v1alpha1.CanaryStep{
				{Pause: &v1alpha1.CanaryPause{Duration: func() *int32 { a := int32(1); return &a }()}},
				{Pause: &v1alpha1.CanaryPause{Duration: func() *int32 { a := int32(2); return &a }()}},
			},
		},
	}
	deploy.Status.UpdateRevision = "2"

	// pod and controller revision is existed.
	f.deployObjects = append(f.deployObjects, deploy)

	pod := newPod(0, deploy.Spec.Template.Labels, true)
	f.podLister = append(f.podLister, pod)

	// for the first sync, it will get object with kube client
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// get the new object for claiming pods
	f.expectGetGameDeploymentAction(deploy.Namespace, deploy.Name)
	// only patch pod, no delete
	f.expectPatchPodAction(pod.Namespace, pod.Name)
	// update game deployment status
	f.expectPatchGameDeploymentSubResourceAction(deploy, nil)
	f.run(types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}.String())
}

// Test delete pod
func TestDeletePod(t *testing.T) {
	f := newFixture(t)
	deploy := test.NewGameDeployment(3)
	f.deployLister = append(f.deployLister, deploy)
	pod := newPod(0, deploy.Spec.Template.Labels, true)
	pod.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
	f.podLister = append(f.podLister, pod)

	d2 := test.NewGameDeployment(3)
	d2.Name = "test-2"
	pod3 := newPod(0, d2.Spec.Template.Labels, true)
	pod3.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(d2, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
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
	expectedKey, _ := controller.KeyFunc(deploy)
	if got, want := key.(string), expectedKey; got != want {
		t.Errorf("queue.Get() = %v, want %v", got, want)
	}

	// pod with not owner reference
	pod2 := test.NewPod(1)
	f.c.deletePod(pod2)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// not pod
	f.c.deletePod(deploy)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}

	// pod with owner reference but deploy not exist
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

func TestEnqueueGameDeploymentForHook(t *testing.T) {
	f := newFixture(t)
	deploy := test.NewGameDeployment(3)
	hr := test.NewHookRunFromTemplate(test.NewHookTemplate(), deploy)
	f.deployLister = append(f.deployLister, deploy)

	f.newController()

	f.c.enqueueGameDeploymentForHook(hr)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	// remove the key from the queue
	for f.c.queue.Len() > 0 {
		f.c.queue.Get()
	}

	a := cache.DeletedFinalStateUnknown{}
	f.c.enqueueGameDeploymentForHook(a)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestAddPod(t *testing.T) {
	f := newFixture(t)

	deploy := test.NewGameDeployment(3)
	f.deployLister = append(f.deployLister, deploy)
	pod := newPod(0, deploy.Spec.Template.Labels, true)
	pod.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
	pod2 := newPod(1, deploy.Spec.Template.Labels, true)

	f.newController()
	f.c.addPod(pod)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Fatalf("queue.Len() = %v, want %v", got, want)
	}
	key, done := f.c.queue.Get()
	if key == nil || done {
		t.Fatalf("failed to enqueue controller for pod %v", pod.Name)
	}
	expectedKey, _ := controller.KeyFunc(deploy)
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

	deploy := test.NewGameDeployment(3)
	deploy2 := test.NewGameDeployment(3)
	deploy2.Name = "test-2"
	deploy3 := test.NewGameDeployment(3)
	deploy3.Name = "test-3"
	deploy4 := test.NewGameDeployment(3)
	deploy4.Name = "test-4"
	f.deployLister = append(f.deployLister, deploy)
	f.deployLister = append(f.deployLister, deploy2)
	f.deployLister = append(f.deployLister, deploy3)
	f.deployLister = append(f.deployLister, deploy4)
	cur := newPod(0, deploy.Spec.Template.Labels, true)
	cur.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
	old := newPod(1, deploy.Spec.Template.Labels, true)
	cur2 := cur.DeepCopy()
	cur2.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy3, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
	cur2.ResourceVersion = "cur2"
	cur2.Name = "cur2"
	cur3 := cur.DeepCopy()
	cur3.OwnerReferences = nil
	cur3.Labels = deploy4.Spec.Template.Labels
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
	old.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy2, v1alpha1.SchemeGroupVersion.WithKind("GameDeployment"))}
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
