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

package hook

import (
	"context"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/providers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-hook-operator/pkg/util/testutil"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookFake "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned/fake"
	hookInformers "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/informers/externalversions"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionfake "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
)

var alwaysReady = func() bool { return true }

type fixture struct {
	t testing.TB
	c *HookController

	kubeClient *fake.Clientset
	apiClient  *apiextensionfake.Clientset
	hookClient *hookFake.Clientset

	hookRunLister []*hookv1alpha1.HookRun

	// informers
	kubeInformer informers.SharedInformerFactory
	hookInformer hookInformers.SharedInformerFactory

	newProvider func(metric hookv1alpha1.Metric) (providers.Provider, error)

	// Actions expected to happen on the client.
	kubeActions []clienttesting.Action
	hookActions []clienttesting.Action

	// Objects from here are also preloaded into NewSimpleFake.
	kubeObjects []runtime.Object
	apiObjects  []runtime.Object
	hookObjects []runtime.Object
}

func assertActions(expect, got []clienttesting.Action, t testing.TB) {
	for i, action := range got {
		if len(expect) < i+1 {
			t.Errorf("%d unexpected actions: %+v", len(got)-len(expect), got[i:])
			break
		}

		expectedAction := expect[i]
		if !(expectedAction.Matches(action.GetVerb(), action.GetResource().Resource) &&
			action.GetSubresource() == expectedAction.GetSubresource()) {
			t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expectedAction, action)
			continue
		}
	}

	if len(expect) > len(got) {
		t.Errorf("%d additional expected actions:%+v", len(expect)-len(got), expect[len(got):])
	}
}

func filterInformerActions(actions []clienttesting.Action) []clienttesting.Action {
	ret := []clienttesting.Action{}
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

func newFixture(t testing.TB) *fixture {
	_ = hookv1alpha1.AddToScheme(scheme.Scheme)
	f := &fixture{}
	f.t = t
	f.kubeObjects = []runtime.Object{}
	f.apiObjects = []runtime.Object{}
	f.hookObjects = []runtime.Object{}
	return f
}

func (f *fixture) newController() {
	// Create the controller
	f.kubeClient = fake.NewSimpleClientset(f.kubeObjects...)
	f.apiClient = apiextensionfake.NewSimpleClientset(f.apiObjects...)
	f.hookClient = hookFake.NewSimpleClientset(f.hookObjects...)
	f.kubeInformer = informers.NewSharedInformerFactory(f.kubeClient, controller.NoResyncPeriodFunc())
	f.hookInformer = hookInformers.NewSharedInformerFactory(f.hookClient, controller.NoResyncPeriodFunc())

	c := NewHookController(f.kubeClient, f.apiClient,
		f.hookClient, f.hookInformer.Tkex().V1alpha1().HookRuns(), &record.FakeRecorder{})
	c.hookRunSynced = alwaysReady

	for _, run := range f.hookRunLister {
		f.hookInformer.Tkex().V1alpha1().HookRuns().Informer().GetIndexer().Add(run)
	}

	f.c = c
}

func (f *fixture) run(hookRunName string) error {
	f.newController()

	err := f.c.sync(hookRunName)
	// TODO: This client-go isn't support merge-patch+json PatchType yet, we need to upgrade it.
	if err != nil && err.Error() != "PatchType is not supported" {
		return err
	}

	assertActions(f.kubeActions, filterInformerActions(f.kubeClient.Actions()), f.t)
	assertActions(f.hookActions, filterInformerActions(f.hookClient.Actions()), f.t)
	return nil
}

func clearQueue(queue workqueue.RateLimitingInterface) {
	for queue.Len() > 0 {
		key, shutdown := queue.Get()
		if shutdown {
			return
		}
		queue.Done(key)
	}
}

func (f *fixture) expectedPatchHookRunSubStatus(namespace, name string, pt types.PatchType, patch []byte) {
	action := clienttesting.NewPatchSubresourceAction(hookv1alpha1.SchemeGroupVersion.WithResource("hookruns"),
		namespace, name, pt, patch, "status")
	f.hookActions = append(f.hookActions, action)
}

func TestEnqueueHookRun(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	f.newController()

	f.c.enqueueHookRun(hr)
	if got, want := f.c.queue.Len(), 1; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	clearQueue(f.c.queue)

	// test error
	f.c.enqueueHookRun(nil)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestEnqueueHookRunAfter(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	f.newController()

	f.c.enqueueHookRunAfter(hr, time.Minute)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
	clearQueue(f.c.queue)

	// test error
	f.c.enqueueHookRunAfter(nil, time.Minute)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestSyncNotFound(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	f.hookRunLister = append(f.hookRunLister, hr)

	// test not found
	f.run(hr.Name)
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestSyncDeletedHookRun(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	hr.DeletionTimestamp = &metav1.Time{Time: time.Now()}
	f.hookRunLister = append(f.hookRunLister, hr)

	f.run(types.NamespacedName{Name: hr.Name, Namespace: hr.Namespace}.String())
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestSyncWithWrongKey(t *testing.T) {
	f := newFixture(t)

	key := "a/b/c"
	expectedError := fmt.Errorf("unexpected key format: %q", key)
	err := f.run(key)
	if !reflect.DeepEqual(err, expectedError) {
		t.Errorf("Expected error %v, got %v", expectedError, err)
	}
}

func TestSyncWithHostIPArgs(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	hr.Spec.Args = append(hr.Spec.Args, hookv1alpha1.Argument{
		Name: "HostIP", Value: func() *string { s := "1.1.1.1"; return &s }()})
	f.hookRunLister = append(f.hookRunLister, hr)

	f.run(types.NamespacedName{Name: hr.Name, Namespace: hr.Namespace}.String())
	if got, want := f.c.queue.Len(), 0; got != want {
		t.Errorf("queue.Len() = %v, want %v", got, want)
	}
}

func TestSync(t *testing.T) {
	f := newFixture(t)
	hr := testutil.NewHookRun("hr1")
	f.hookRunLister = append(f.hookRunLister, hr)

	f.expectedPatchHookRunSubStatus(hr.Namespace, hr.Name, types.MergePatchType, nil)

	f.run(types.NamespacedName{Name: hr.Name, Namespace: hr.Namespace}.String())
}

func buildVersionClientV1beta1(t *testing.T) (kubernetes.Interface, apiextension.Interface) {
	f := newFixture(t)
	f.newController()
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bcs-hook-operator",
			Namespace: "bcs-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "bcs-hook-operator:v0.0",
						},
					},
				},
			},
		},
	}
	f.kubeClient.AppsV1().DaemonSets("bcs-system").Create(context.TODO(), &ds, metav1.CreateOptions{})
	templatev1beta1 := apiv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hooktemplates.tkex.tencent.com",
			Annotations: map[string]string{
				"version": "v0.0",
			},
		},
	}
	f.apiClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(),
		&templatev1beta1, metav1.CreateOptions{})
	runv1beta1 := apiv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hookruns.tkex.tencent.com",
			Annotations: map[string]string{
				"version": "v0.0",
			},
		},
	}
	f.apiClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(),
		&runv1beta1, metav1.CreateOptions{})
	return f.kubeClient, f.apiClient
}

func buildVersionClientV1(t *testing.T) (kubernetes.Interface, apiextension.Interface) {
	f := newFixture(t)
	f.newController()
	ds := appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bcs-hook-operator",
			Namespace: "bcs-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "bcs-hook-operator:v0.1",
						},
					},
				},
			},
		},
	}
	f.kubeClient.AppsV1().DaemonSets("bcs-system").Create(context.TODO(), &ds, metav1.CreateOptions{})
	templatev1 := apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hooktemplates.tkex.tencent.com",
			Annotations: map[string]string{
				"version": "v0.1",
			},
		},
	}
	f.apiClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(),
		&templatev1, metav1.CreateOptions{})
	runv1 := apiv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hookruns.tkex.tencent.com",
			Annotations: map[string]string{
				"version": "v0.1",
			},
		},
	}
	f.apiClient.ApiextensionsV1().CustomResourceDefinitions().Create(context.TODO(),
		&runv1, metav1.CreateOptions{})
	return f.kubeClient, f.apiClient
}
func TestHookController_getVersion(t *testing.T) {

	type fields struct {
		kubeClient         kubernetes.Interface
		apiextensionClient apiextension.Interface
	}
	tests := []struct {
		name                    string
		fields                  fields
		wantImageVersion        string
		wantHookrunVersion      string
		wantHooktemplateVersion string
	}{
		{
			name: "version = v0.0, with v1beta1 crd",
			fields: fields{
				kubeClient: func() kubernetes.Interface {
					tmp, _ := buildVersionClientV1beta1(t)
					return tmp
				}(),
				apiextensionClient: func() apiextension.Interface {
					_, tmp := buildVersionClientV1beta1(t)
					return tmp
				}(),
			},
			wantImageVersion:        "v0.0",
			wantHookrunVersion:      "v1beta1-v0.0",
			wantHooktemplateVersion: "v1beta1-v0.0",
		},
		{
			name: "version = v0.1, with v1 crd",
			fields: fields{
				kubeClient: func() kubernetes.Interface {
					tmp, _ := buildVersionClientV1(t)
					return tmp
				}(),
				apiextensionClient: func() apiextension.Interface {
					_, tmp := buildVersionClientV1(t)
					return tmp
				}(),
			},
			wantImageVersion:        "v0.1",
			wantHookrunVersion:      "v1-v0.1",
			wantHooktemplateVersion: "v1-v0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hc := &HookController{
				kubeClient:         tt.fields.kubeClient,
				apiextensionClient: tt.fields.apiextensionClient,
			}
			gotImageVersion, gotHookrunVersion, gotHooktemplateVersion := hc.getVersion()
			if gotImageVersion != tt.wantImageVersion {
				t.Errorf("HookController.getVersion() gotImageVersion = %v, want %v", gotImageVersion, tt.wantImageVersion)
			}
			if gotHookrunVersion != tt.wantHookrunVersion {
				t.Errorf("HookController.getVersion() gotHookrunVersion = %v, want %v", gotHookrunVersion, tt.wantHookrunVersion)
			}
			if gotHooktemplateVersion != tt.wantHooktemplateVersion {
				t.Errorf("HookController.getVersion() gotHooktemplateVersion = %v, want %v", gotHooktemplateVersion,
					tt.wantHooktemplateVersion)
			}
		})
	}
}
