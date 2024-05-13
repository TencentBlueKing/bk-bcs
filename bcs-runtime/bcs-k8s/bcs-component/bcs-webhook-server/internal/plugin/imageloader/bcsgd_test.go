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

package imageloader

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/apis/tkex/v1alpha1"
	tkexFake "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/clientset/versioned/fake"
	tkexinformers "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/informers/externalversions"

	"k8s.io/api/admission/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller"
)

var (
	fakeKubeClient = k8sfake.NewSimpleClientset()
	fakeTkexclient = tkexFake.NewSimpleClientset()
)

func expectCreateJobAction(j *batchv1.Job) core.CreateActionImpl {
	return core.NewCreateAction(batchv1.SchemeGroupVersion.WithResource("jobs"), apiv1.NamespaceDefault, j)
}

func exceptGetJobAction() core.GetActionImpl {
	return core.NewGetAction(batchv1.SchemeGroupVersion.WithResource("jobs"), apiv1.NamespaceDefault, "test-gd")
}

func expectDeleteJobAction(name string) core.DeleteActionImpl {
	return core.NewDeleteAction(batchv1.SchemeGroupVersion.WithResource("jobs"), apiv1.NamespaceDefault, name)
}

func expectListPodsAction() core.ListActionImpl {
	return core.NewListAction(schema.GroupVersionResource{
		Version: "v1", Resource: "pods",
	}, apiv1.SchemeGroupVersion.WithKind("Pods"), apiv1.NamespaceDefault, metav1.ListOptions{
		LabelSelector: "app=test-gd",
	})
}

func expectCreateEventAction() core.CreateActionImpl {
	return core.NewCreateAction(apiv1.SchemeGroupVersion.WithResource("events"), apiv1.NamespaceDefault, &apiv1.Event{})
}

// NOCC:tosa/fn_length(设计如此)
func Test_bcsgdWorkload_LoadImageBeforeUpdate(t *testing.T) {
	blog.InitLogs(conf.LogConfig{Verbosity: 4, ToStdErr: true, AlsoToStdErr: true})
	b := newBcsGdWorkload()

	type args struct {
		ar v1beta1.AdmissionReview
	}
	tests := []struct {
		name string
		// fields fields
		args               args
		want               *v1beta1.AdmissionResponse
		expectedKubeAction []core.Action
		expectedTkexAction []core.Action
	}{
		{
			name: "only update annotations",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations = map[string]string{
									"a": "c",
								}
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
					},
				},
			},
			want: func() *v1beta1.AdmissionResponse {
				return toAdmissionResponse(nil)
			}(),
			// expectedKubeAction: []core.Action{expectListPodsAction()},
		},
		{
			name: "firstly update images",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Spec.Template.Spec.Containers[0].Image = "nginx:v2"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
					},
				},
			},
			want: func() *v1beta1.AdmissionResponse {
				finalPatch := newPatch("nginx:v2", "nginx:v1")
				return toAdmissionResponse(nil, finalPatch)
			}(),
			expectedKubeAction: []core.Action{
				expectListPodsAction(),
				expectCreateEventAction(),
				exceptGetJobAction(),
				expectCreateJobAction(newFakeJob(b, "nginx:v2")),
			},
		},
		{
			name: "job have done",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								gd.Labels[imagePreloadDoneLabel] = "true"
								gd.Spec.Template.Spec.Containers[0].Image = "nginx:v2"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
					},
				},
			},
			want: func() *v1beta1.AdmissionResponse {
				anno := fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}", imageUpdateAnno)
				label := fmt.Sprintf("{\"op\":\"remove\",\"path\":\"/metadata/labels/%s\"}", imagePreloadDoneLabel)
				finalPatch := fmt.Sprintf("[%s]", strings.Join([]string{anno, label}, ","))
				return toAdmissionResponse(nil, finalPatch)
			}(),
			expectedKubeAction: []core.Action{
				expectDeleteJobAction("gamedeployment-default-test-gd"),
			},
		},
		{
			name: "update same container when job is running",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								gd.Spec.Template.Spec.Containers[0].Image = "nginx:v3"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
					},
				},
			},
			want: func() *v1beta1.AdmissionResponse {
				finalPatch := newPatch("nginx:v3", "nginx:v1")
				return toAdmissionResponse(nil, finalPatch)
			}(),
			expectedKubeAction: []core.Action{
				expectListPodsAction(),
				expectCreateEventAction(),
				exceptGetJobAction(),
				expectCreateJobAction(newFakeJob(b, "nginx:v3")),
			},
		},
		{
			name: "change to update second container when job of first container is running",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGD()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								gd.Spec.Template.Spec.Containers[1].Image = "web:v2"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
					},
				},
			},
			want: func() *v1beta1.AdmissionResponse {
				imagePatch := fmt.Sprintf(
					"{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/1/image\",\"value\":\"%s\"}", "web:v1")
				storePatch := fmt.Sprintf(
					"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/1/image\",\"value\":\"%s\"}]", "web:v2")
				annoPatch := fmt.Sprintf("{\"op\":\"add\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}",
					imageUpdateAnno, strings.ReplaceAll(storePatch, "\"", "\\\""))
				finalPatch := fmt.Sprintf("[%s]", strings.Join([]string{imagePatch, annoPatch}, ","))
				return toAdmissionResponse(nil, finalPatch)
			}(),
			expectedKubeAction: []core.Action{
				expectListPodsAction(),
				expectCreateEventAction(),
				exceptGetJobAction(),
				expectDeleteJobAction("gamedeployment-default-test-gd"),
				expectCreateJobAction(newFakeJob(b, "web:v2")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeKubeClient.ClearActions()
			fakeTkexclient.ClearActions()

			if got := b.LoadImageBeforeUpdate(tt.args.ar); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bcsgdWorkload.LoadImageBeforeUpdate() = %v,\n want %v", got, tt.want)
				t.Logf("got = %v", string(got.Patch))
				t.Logf("want = %v", string(tt.want.Patch))
			}
			assertActions(tt.expectedKubeAction, fakeKubeClient.Actions(), t)
			assertActions(tt.expectedTkexAction, fakeTkexclient.Actions(), t)

		})
	}

}

func assertActions(expect, got []core.Action, t testing.TB) {
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

func newPatch(new, old string) string {
	imagePatch := fmt.Sprintf(
		"{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"%s\"}", old)
	storePatch := fmt.Sprintf(
		"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"%s\"}]", new)
	annoPatch := fmt.Sprintf("{\"op\":\"add\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}",
		imageUpdateAnno, strings.ReplaceAll(storePatch, "\"", "\\\""))
	finalPatch := fmt.Sprintf("[%s]", strings.Join([]string{imagePatch, annoPatch}, ","))
	return finalPatch
}

func newFakeJob(b *bcsgdWorkload, image string) *batchv1.Job {
	containers := []apiv1.Container{
		{
			Name:            "container-A",
			Image:           image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command:         []string{"echo", "pull " + image},
		},
	}
	job, _ := b.generateJobByDiff(newGD(), containers)
	return job
}

func newGD() *tkexv1alpha1.GameDeployment {
	gd := tkexv1alpha1.GameDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gd",
			Namespace: apiv1.NamespaceDefault,
			Annotations: map[string]string{
				"a": "b",
			},
			Labels: map[string]string{},
		},
		Spec: tkexv1alpha1.GameDeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-gd",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"a": "b",
					},
					Labels: map[string]string{
						"app": "test-gd",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "container-A",
							Image: "nginx:v1",
						},
						{
							Name:  "container-B",
							Image: "web:v1",
						},
					},
				},
			},
			UpdateStrategy: tkexv1alpha1.GameDeploymentUpdateStrategy{
				Type: tkexv1alpha1.InPlaceGameDeploymentUpdateStrategyType,
			},
		},
	}
	return &gd
}

func newBcsGdWorkload() *bcsgdWorkload {
	workloadName := metav1.GroupVersionKind{
		Group:   tkexv1alpha1.GroupVersion.Group,
		Version: tkexv1alpha1.GroupVersion.Version,
		Kind:    tkexv1alpha1.KindGameDeployment,
	}.String()

	tkexClient := fakeTkexclient
	tkexInformers := tkexinformers.NewSharedInformerFactory(tkexClient, controller.NoResyncPeriodFunc())
	tkexStop := make(chan struct{})
	defer close(tkexStop)
	tkexInformers.Start(tkexStop)
	tkexInformers.WaitForCacheSync(tkexStop)

	fakeGD := newGD()
	_, _ = tkexClient.TkexV1alpha1().GameDeployments(apiv1.NamespaceDefault).Create(
		context.TODO(), fakeGD, metav1.CreateOptions{})
	err := tkexInformers.Tkex().V1alpha1().GameDeployments().Informer().GetIndexer().Add(fakeGD)
	if err != nil {
		fmt.Printf("informer error: %+v", err)
	}
	tkexClient.ClearActions()
	bcsworkload := bcsgdWorkload{
		name:     workloadName,
		client:   tkexClient,
		informer: tkexInformers.Tkex().V1alpha1().GameDeployments().Informer(),
		lister:   tkexInformers.Tkex().V1alpha1().GameDeployments().Lister(),
		i:        newImageLoader(),
	}
	return &bcsworkload
}

func newImageLoader() *imageLoader {
	kubeClient := fakeKubeClient
	kubeInformers := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
	kubeStop := make(chan struct{})
	defer close(kubeStop)
	kubeInformers.Start(kubeStop)
	kubeInformers.WaitForCacheSync(kubeStop)

	// mock nodes objects
	fakeNodes := []*apiv1.Node{
		func() *apiv1.Node {
			node := apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
			}
			return &node
		}(),
		func() *apiv1.Node {
			node := apiv1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
			}
			return &node
		}(),
	}
	for i := range fakeNodes {
		_, _ = kubeClient.CoreV1().Nodes().Create(context.TODO(), fakeNodes[i], metav1.CreateOptions{})
		err := kubeInformers.Core().V1().Nodes().Informer().GetIndexer().Add(fakeNodes[i])
		if err != nil {
			fmt.Printf("informer error: %+v", err)
		}
	}
	nodeLister := kubeInformers.Core().V1().Nodes().Lister()

	// mock pod object
	fakePods := []*apiv1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gd-a",
				Namespace: apiv1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test-gd",
				},
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "container-A",
						Image: "nginx:v1",
					},
				},
				NodeName: "node-1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gd-b",
				Namespace: apiv1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test-gd",
				},
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "container-A",
						Image: "nginx:v1",
					},
				},
				NodeName: "node-1",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gd-c",
				Namespace: apiv1.NamespaceDefault,
				Labels: map[string]string{
					"app": "test-gd",
				},
			},
			Spec: apiv1.PodSpec{
				Containers: []apiv1.Container{
					{
						Name:  "container-A",
						Image: "nginx:v1",
					},
				},
				NodeName: "node-2",
			},
		},
	}
	for i := range fakePods {
		_, _ = kubeClient.CoreV1().Pods(apiv1.NamespaceDefault).Create(context.TODO(), fakePods[i], metav1.CreateOptions{})
		err := kubeInformers.Core().V1().Pods().Informer().GetIndexer().Add(fakePods[i])
		if err != nil {
			fmt.Printf("informer error: %+v", err)
		}
	}

	kubeClient.ClearActions()

	fakeImageLoader := imageLoader{
		k8sClient:  kubeClient,
		nodeLister: nodeLister,
		jobLister:  kubeInformers.Batch().V1().Jobs().Lister(),
	}
	return &fakeImageLoader
}
