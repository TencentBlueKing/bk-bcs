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
	stsplus "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	apps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"testing"
)

func expectCreatePodAction(pod *corev1.Pod) testing2.Action {
	return testing2.NewCreateAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		pod.Namespace, pod)
}

func expectGetPodAction(namespace, name string) testing2.Action {
	return testing2.NewGetAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name)
}

func expectUpdatePodAction(namespace string, object runtime.Object) testing2.Action {
	return testing2.NewUpdateAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, object)
}

func expectDeletePodAction(namespace, name string) testing2.Action {
	return testing2.NewDeleteAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name)
}

func expectPatchPodAction(namespace, name string, patchType types.PatchType) testing2.Action {
	return testing2.NewPatchAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "pods"},
		namespace, name, patchType, []byte{})
}

func expectCreatePVCAction(pvc *corev1.PersistentVolumeClaim) testing2.Action {
	return testing2.NewCreateAction(schema.GroupVersionResource{Version: corev1.SchemeGroupVersion.Version, Resource: "persistentvolumeclaims"},
		pvc.Namespace, pvc)
}

func expectCreateControllerRevisions(cr *apps.ControllerRevision) testing2.Action {
	return testing2.NewCreateAction(schema.GroupVersionResource{Group: apps.SchemeGroupVersion.Group,
		Version: apps.SchemeGroupVersion.Version, Resource: "controllerrevisions"},
		cr.Namespace, cr)
}

func TestCreateGameStatefulSetPod(t *testing.T) {
	tests := []struct {
		name            string
		set             *stsplus.GameStatefulSet
		pod             *corev1.Pod
		pvcLister       []*corev1.PersistentVolumeClaim
		podLister       []*corev1.Pod
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "with pv template",
			set:  newStatefulSet(1, true),
			pod:  newPod(1, nil, true),
			expectedActions: []testing2.Action{
				expectCreatePVCAction(newPVC("data-pod-1", map[string]string{"foo": "bar"})),
				expectCreatePodAction(newPod(1, nil, true)),
			},
			expectedError: nil,
		},
		{
			name:      "pod is existed",
			set:       newStatefulSet(1, true),
			pod:       newPod(1, nil, true),
			podLister: []*corev1.Pod{newPod(1, nil, true)},
			pvcLister: []*corev1.PersistentVolumeClaim{
				newPVC("data-foo-1", nil),
				newPVC("log-foo-1", nil),
			},
			expectedError: k8serrors.NewAlreadyExists(schema.GroupResource{Resource: "pods"}, "foo-1"),
			expectedActions: []testing2.Action{
				expectCreatePodAction(newPod(1, nil, true)),
			},
		},
		{
			name: "without pvc template",
			set:  testutil.NewGameStatefulSet(1),
			pod:  newPod(1, nil, true),
			expectedActions: []testing2.Action{
				expectCreatePodAction(newPod(1, nil, true)),
			},
			expectedError: nil,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			spc := realGameStatefulSetPodControl{
				client:    client,
				pvcLister: informerFactory.Core().V1().PersistentVolumeClaims().Lister(),
				recorder:  record.NewFakeRecorder(100),
				metrics:   newMetrics(),
			}
			for _, pod := range s.podLister {
				client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			}
			for _, claim := range s.pvcLister {
				informerFactory.Core().V1().PersistentVolumeClaims().Informer().GetIndexer().Add(claim)
			}
			client.ClearActions()
			err := spc.CreateGameStatefulSetPod(s.set, s.pod)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("Unexpected error: %v, expected: %v", err, s.expectedError)
			}
			actions := testutil.FilterActions(client.Actions(), testutil.FilterCreateAction, testutil.FilterUpdateAction)
			expectActions := testutil.FilterActions(s.expectedActions, testutil.FilterCreateAction, testutil.FilterUpdateAction)
			if !testutil.EqualActions(expectActions, actions) {
				t.Errorf("Unexpected actions: \n\t%v\nexpected \n\t%v", actions, expectActions)
			}
		})
	}
}

func TestUpdateGameStatefulSetPod(t *testing.T) {
	tests := []struct {
		name            string
		set             *stsplus.GameStatefulSet
		pod             *corev1.Pod
		podLister       []*corev1.Pod
		podClient       []*corev1.Pod
		expectedActions []testing2.Action
		expectedError   error
	}{
		{
			name: "pod info inconsistent",
			set:  testutil.NewGameStatefulSet(1),
			pod:  newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "bar-1"}, true),
			podClient: []*corev1.Pod{
				newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			},
			expectedActions: []testing2.Action{
				expectUpdatePodAction(corev1.NamespaceDefault, newPod(1, nil, true)),
			},
			expectedError: nil,
		},
		{
			name: "pod storage inconsistent",
			set:  newStatefulSet(1, true),
			pod:  newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			podClient: []*corev1.Pod{
				newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			},
			expectedActions: []testing2.Action{
				expectCreatePVCAction(newPVC("data-pod-1", map[string]string{"foo": "bar"})),
				expectUpdatePodAction(corev1.NamespaceDefault, newPod(1, nil, true)),
			},
			expectedError: nil,
		},
		{
			name: "pod is consistent",
			set:  testutil.NewGameStatefulSet(1),
			pod:  newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			podClient: []*corev1.Pod{
				newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			},
			expectedError: nil,
		},
		{
			name: "pod update fails, get new one and retry",
			set:  testutil.NewGameStatefulSet(1),
			pod:  newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "bar-1"}, true),
			expectedActions: []testing2.Action{
				expectUpdatePodAction(corev1.NamespaceDefault, newPod(1, nil, true)),
			},
			podLister: []*corev1.Pod{
				newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "foo-1"}, true),
			},
			expectedError: k8serrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "foo-1"),
		},
		{
			name: "pod update fails, get new one and retry fail",
			set:  testutil.NewGameStatefulSet(1),
			pod:  newPod(1, map[string]string{stsplus.GameStatefulSetPodNameLabel: "bar-1"}, true),
			expectedActions: []testing2.Action{
				expectUpdatePodAction(corev1.NamespaceDefault, newPod(1, nil, true)),
			},
			expectedError: k8serrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "foo-1"),
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			spc := realGameStatefulSetPodControl{
				client:    client,
				pvcLister: informerFactory.Core().V1().PersistentVolumeClaims().Lister(),
				podLister: informerFactory.Core().V1().Pods().Lister(),
				recorder:  record.NewFakeRecorder(100),
			}
			for _, pod := range s.podClient {
				client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			}
			for _, pod := range s.podLister {
				informerFactory.Core().V1().Pods().Informer().GetIndexer().Add(pod)
			}
			client.ClearActions()
			err := spc.UpdateGameStatefulSetPod(s.set, s.pod)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("Unexpected error: %v, expected: %v", err, s.expectedError)
			}
			actions := testutil.FilterActions(client.Actions(), testutil.FilterCreateAction, testutil.FilterUpdateAction)
			expectActions := testutil.FilterActions(s.expectedActions, testutil.FilterCreateAction, testutil.FilterUpdateAction)
			if !testutil.EqualActions(expectActions, actions) {
				t.Errorf("Unexpected actions: \n\t%v\nexpected \n\t%v", actions, expectActions)
			}
		})
	}
}

func newPodWithAnnotations(annotations map[string]string, nodeName string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "foo-1",
			Namespace:   corev1.NamespaceDefault,
			Annotations: annotations,
		},
		Spec: corev1.PodSpec{
			NodeName: nodeName,
		},
	}
}

func newNode(ready bool) *corev1.Node {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
	}
	if ready {
		node.Status.Conditions = append(node.Status.Conditions, corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionTrue})
	} else {
		node.Status.Conditions = append(node.Status.Conditions, corev1.NodeCondition{Type: corev1.NodeReady, Status: corev1.ConditionFalse})
	}
	return node
}

func TestForceDeleteGameStatefulSetPod(t *testing.T) {
	tests := []struct {
		name            string
		set             *stsplus.GameStatefulSet
		pod             *corev1.Pod
		nodeLister      []*corev1.Node
		podClient       []*corev1.Pod
		expectedActions []testing2.Action
		expectedError   error
		expectedDelete  bool
	}{
		{
			name:           "don't have delete label, no force delete",
			set:            newStatefulSet(1, true),
			pod:            newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "false"}, "foo"),
			expectedDelete: false,
		},
		{
			name:           "node is not found, deleted failed",
			set:            newStatefulSet(1, true),
			pod:            newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "true"}, "foo"),
			expectedError:  k8serrors.NewNotFound(schema.GroupResource{Resource: "node"}, "foo"),
			expectedDelete: false,
		},
		{
			name: "node is not ready, force delete",
			set:  newStatefulSet(1, true),
			pod:  newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "true"}, "foo"),
			podClient: []*corev1.Pod{
				newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "true"}, "foo"),
			},
			nodeLister:     []*corev1.Node{newNode(false)},
			expectedDelete: true,
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name:           "node is not ready, force delete failed because of pod is not found",
			set:            newStatefulSet(1, true),
			pod:            newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "true"}, "foo"),
			nodeLister:     []*corev1.Node{newNode(false)},
			expectedError:  k8serrors.NewNotFound(schema.GroupResource{Resource: "pods"}, "foo-1"),
			expectedDelete: false,
			expectedActions: []testing2.Action{
				expectDeletePodAction(corev1.NamespaceDefault, "foo-1"),
			},
		},
		{
			name:           "node is ready, no force delete",
			set:            newStatefulSet(1, true),
			pod:            newPodWithAnnotations(map[string]string{podNodeLostForceDeleteKey: "true"}, "foo"),
			nodeLister:     []*corev1.Node{newNode(true)},
			expectedDelete: false,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			client := fake.NewSimpleClientset()
			informerFactory := informers.NewSharedInformerFactory(client, controller.NoResyncPeriodFunc())
			spc := realGameStatefulSetPodControl{
				client:     client,
				nodeLister: informerFactory.Core().V1().Nodes().Lister(),
				podLister:  informerFactory.Core().V1().Pods().Lister(),
				recorder:   record.NewFakeRecorder(100),
			}
			for _, pod := range s.podClient {
				client.CoreV1().Pods(pod.Namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			}
			for _, node := range s.nodeLister {
				informerFactory.Core().V1().Nodes().Informer().GetIndexer().Add(node)
			}
			client.ClearActions()
			deleted, err := spc.ForceDeleteGameStatefulSetPod(s.set, s.pod)
			if !reflect.DeepEqual(err, s.expectedError) {
				t.Errorf("Unexpected error: %v, expected: %v", err, s.expectedError)
			}
			if deleted != s.expectedDelete {
				t.Errorf("Unexpected deleted: %v, expected: %v", deleted, s.expectedDelete)
			}
			actions := testutil.FilterActions(client.Actions(), testutil.FilterCreateAction, testutil.FilterUpdateAction)
			expectActions := testutil.FilterActions(s.expectedActions, testutil.FilterCreateAction, testutil.FilterUpdateAction)
			if !testutil.EqualActions(expectActions, actions) {
				t.Errorf("Unexpected actions: \n\t%v\nexpected \n\t%v", actions, expectActions)
			}
		})
	}
}
