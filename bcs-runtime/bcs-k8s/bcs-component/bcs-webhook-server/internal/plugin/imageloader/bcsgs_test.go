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
	tkexinformers "github.com/Tencent/bk-bcs/bcs-scenarios/kourse/pkg/client/informers/externalversions"

	"k8s.io/api/admission/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller"
)

// NOCC:tosa/fn_length(设计如此)
func Test_bcsgsWorkload_LoadImageBeforeUpdate(t *testing.T) {
	blog.InitLogs(conf.LogConfig{Verbosity: 4, ToStdErr: true, AlsoToStdErr: true})
	b := newBcsGsWorkload()

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
								gd := newGS()
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
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
								gd := newGS()
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
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
				expectCreateJobAction(newFakeJob2(b, "nginx:v2")),
			},
		},
		{
			name: "job have done",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
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
				expectDeleteJobAction("GameStatefulSet-default-test-gd"),
			},
		},
		{
			name: "update same container when job is running",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
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
				expectCreateJobAction(newFakeJob2(b, "nginx:v3")),
			},
		},
		{
			name: "change to update second container when job of first container is running",
			args: args{
				ar: v1beta1.AdmissionReview{
					Request: &v1beta1.AdmissionRequest{
						OldObject: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
								gd.Annotations[imageUpdateAnno] =
									"[{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/0/image\",\"value\":\"nginx:v2\"}]"
								js, _ := runtime.Encode(bcsCodec, gd)
								return js
							}(),
						},
						Object: runtime.RawExtension{
							Raw: func() []byte {
								gd := newGS()
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
				expectDeleteJobAction("GameStatefulSet-default-test-gd"),
				expectCreateJobAction(newFakeJob2(b, "web:v2")),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeKubeClient.ClearActions()
			fakeTkexclient.ClearActions()

			if got := b.LoadImageBeforeUpdate(tt.args.ar); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("bcsgsWorkload.LoadImageBeforeUpdate() = %v,\n want %v", got, tt.want)
				t.Logf("got = %v", string(got.Patch))
			}
			assertActions(tt.expectedKubeAction, fakeKubeClient.Actions(), t)
			assertActions(tt.expectedTkexAction, fakeTkexclient.Actions(), t)

		})
	}

}

func newFakeJob2(b *bcsgsWorkload, image string) *batchv1.Job {
	containers := []apiv1.Container{
		{
			Name:            "container-A",
			Image:           image,
			ImagePullPolicy: apiv1.PullIfNotPresent,
			Command:         []string{"echo", "pull " + image},
		},
	}
	job, _ := b.generateJobByDiff(newGS(), containers)
	return job
}

func newGS() *tkexv1alpha1.GameStatefulSet {
	gd := tkexv1alpha1.GameStatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-gd",
			Namespace: apiv1.NamespaceDefault,
			Annotations: map[string]string{
				"a": "b",
			},
			Labels: map[string]string{},
		},
		Spec: tkexv1alpha1.GameStatefulSetSpec{
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
			UpdateStrategy: tkexv1alpha1.GameStatefulSetUpdateStrategy{
				Type: tkexv1alpha1.InplaceUpdateGameStatefulSetStrategyType,
			},
		},
	}
	return &gd
}

func newBcsGsWorkload() *bcsgsWorkload {
	workloadName := metav1.GroupVersionKind{
		Group:   tkexv1alpha1.GroupVersion.Group,
		Version: tkexv1alpha1.GroupVersion.Version,
		Kind:    tkexv1alpha1.KindGameStatefulSet,
	}.String()

	tkexClient := fakeTkexclient
	tkexInformers := tkexinformers.NewSharedInformerFactory(tkexClient, controller.NoResyncPeriodFunc())
	tkexStop := make(chan struct{})
	defer close(tkexStop)
	tkexInformers.Start(tkexStop)
	tkexInformers.WaitForCacheSync(tkexStop)

	fakeGS := newGS()
	_, _ = tkexClient.TkexV1alpha1().GameStatefulSets(apiv1.NamespaceDefault).Create(
		context.TODO(), fakeGS, metav1.CreateOptions{})
	err := tkexInformers.Tkex().V1alpha1().GameStatefulSets().Informer().GetIndexer().Add(fakeGS)
	if err != nil {
		fmt.Printf("informer error: %+v", err)
	}
	tkexClient.ClearActions()
	bcsworkload := bcsgsWorkload{
		name:     workloadName,
		client:   tkexClient,
		informer: tkexInformers.Tkex().V1alpha1().GameStatefulSets().Informer(),
		lister:   tkexInformers.Tkex().V1alpha1().GameStatefulSets().Lister(),
		i:        newImageLoader(),
	}
	return &bcsworkload
}
