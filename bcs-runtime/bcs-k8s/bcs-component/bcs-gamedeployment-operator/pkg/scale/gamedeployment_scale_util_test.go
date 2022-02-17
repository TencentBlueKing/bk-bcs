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

package scale

import (
	//"context"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/test"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/controller"
)

func newFakeNodes() corelisters.NodeLister {
	// init kube controller
	kubeClient := fake.NewSimpleClientset()
	kubeInformers := informers.NewSharedInformerFactory(kubeClient, controller.NoResyncPeriodFunc())
	kubeStop := make(chan struct{})
	defer close(kubeStop)
	kubeInformers.Start(kubeStop)
	kubeInformers.WaitForCacheSync(kubeStop)

	fakeNodes := []*corev1.Node{
		func() *corev1.Node {
			node := corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-cost-20",
					Annotations: map[string]string{
						NodeDeletionCost: "20",
					},
				},
			}
			return &node
		}(),
		func() *corev1.Node {
			node := corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-cost-10",
					Annotations: map[string]string{
						NodeDeletionCost: "10",
					},
				},
			}
			return &node
		}(),
		func() *corev1.Node {
			node := corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-cost-ff",
					Annotations: map[string]string{
						NodeDeletionCost: "ff",
					},
				},
			}
			return &node
		}(),
		func() *corev1.Node {
			node := corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-without-cost",
				},
			}
			return &node
		}(),
	}

	// mock nodes objects
	for i := range fakeNodes {
		//_, _ = kubeClient.CoreV1().Nodes().Create(context.TODO(), &fakeNodes[i], metav1.CreateOptions{})
		err := kubeInformers.Core().V1().Nodes().Informer().GetIndexer().Add(fakeNodes[i])
		if err != nil {
			fmt.Printf("informer error: %+v", err)
		}
	}
	return kubeInformers.Core().V1().Nodes().Lister()
}

func TestGenAvailableIDs(t *testing.T) {
	pod1 := test.NewPod(1)
	pod1.Labels[v1alpha1.GameDeploymentInstanceID] = "1"
	pod2 := test.NewPod(2)
	pod2.Labels[v1alpha1.GameDeploymentInstanceID] = "2"
	pod3 := test.NewPod(3)
	pod3.Labels[v1alpha1.GameDeploymentInstanceID] = "3"
	got := genAvailableIDs(3, []*corev1.Pod{pod1, pod2, pod3})
	if got.Len() != 3 {
		t.Errorf("expected 3, got %d", got.Len())
	}
	for _, s := range got.List() {
		if len(s) != LengthOfInstanceID {
			t.Errorf("got %d, want %d", len(s), LengthOfInstanceID)
		}
	}
}

func TestCalculateDiffs(t *testing.T) {
	tests := []struct {
		name              string
		deploy            *v1alpha1.GameDeployment
		revConsistent     bool
		totalPod          int
		notUpdatedPods    int
		expectedTotalDiff int
		expectedRevDiff   int
	}{
		{
			name:              "revision is consistent",
			deploy:            test.NewGameDeployment(1),
			revConsistent:     true,
			totalPod:          1,
			notUpdatedPods:    1,
			expectedTotalDiff: 0,
			expectedRevDiff:   0,
		},
		{
			name: "partition is specified",
			deploy: func() *v1alpha1.GameDeployment {
				d := test.NewGameDeployment(5)
				d.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{Partition: func() *int32 { a := int32(1); return &a }()},
						{Partition: func() *int32 { a := int32(2); return &a }()},
					},
				}
				d.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return d
			}(),
			revConsistent:     false,
			totalPod:          5,
			notUpdatedPods:    3,
			expectedTotalDiff: 0,
			expectedRevDiff:   2,
		},
		{
			name: "partition has not satisfied",
			deploy: func() *v1alpha1.GameDeployment {
				d := test.NewGameDeployment(5)
				d.Spec.UpdateStrategy.CanaryStrategy = &v1alpha1.CanaryStrategy{
					Steps: []v1alpha1.CanaryStep{
						{Partition: func() *int32 { a := int32(1); return &a }()},
						{Partition: func() *int32 { a := int32(2); return &a }()},
					},
				}
				d.Spec.UpdateStrategy.MaxSurge = &intstr.IntOrString{IntVal: 1}
				d.Status.CurrentStepIndex = func() *int32 { a := int32(0); return &a }()
				return d
			}(),
			revConsistent:     false,
			totalPod:          7,
			notUpdatedPods:    3,
			expectedTotalDiff: 1,
			expectedRevDiff:   2,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			totalDiff, revDiff := calculateDiffs(s.deploy, s.revConsistent, s.totalPod, s.notUpdatedPods)
			if totalDiff != s.expectedTotalDiff {
				t.Errorf("totalDiff expected %d, got %d", s.expectedTotalDiff, totalDiff)
			}
			if revDiff != s.expectedRevDiff {
				t.Errorf("revDiff expected %d, got %d", s.expectedRevDiff, revDiff)
			}
		})
	}
}

func TestChoosePodsToDelete(t *testing.T) {
	nodeLister := newFakeNodes()
	tests := []struct {
		name           string
		totalDiff      int
		currentRevDiff int
		notUpdatedPods []*corev1.Pod
		updatedPods    []*corev1.Pod
		sortMethod     string
		expectedPods   []*corev1.Pod
	}{
		{
			name:           "currentRevDiff greater or equal than totalDiff",
			totalDiff:      1,
			currentRevDiff: 3,
			notUpdatedPods: []*corev1.Pod{
				test.NewPod(1),
				test.NewPod(2),
				test.NewPod(3),
				test.NewPod(4),
				test.NewPod(5),
			},
			updatedPods: []*corev1.Pod{},
			expectedPods: []*corev1.Pod{
				test.NewPod(1),
			},
		},
		{
			name:           "currentRevDiff greater or equal than totalDiff, and diff greater than pods number",
			totalDiff:      2,
			currentRevDiff: 3,
			notUpdatedPods: []*corev1.Pod{
				test.NewPod(1),
			},
			updatedPods: []*corev1.Pod{},
			expectedPods: []*corev1.Pod{
				test.NewPod(1),
			},
		},
		{
			name:           "another sort method",
			totalDiff:      1,
			currentRevDiff: 3,
			notUpdatedPods: []*corev1.Pod{
				test.NewPod(1),
				test.NewPod(2),
				test.NewPod(3),
				test.NewPod(4),
				test.NewPod(5),
			},
			sortMethod:  CostSortMethodDescend,
			updatedPods: []*corev1.Pod{},
			expectedPods: []*corev1.Pod{
				test.NewPod(1),
			},
		},
		{
			name:           "currentRevDiff greater than zero, and less than totalDiff",
			totalDiff:      3,
			currentRevDiff: 1,
			notUpdatedPods: []*corev1.Pod{
				test.NewPod(1),
				test.NewPod(2),
			},
			updatedPods: []*corev1.Pod{
				test.NewPod(3),
				test.NewPod(4),
				test.NewPod(5),
			},
			expectedPods: []*corev1.Pod{
				test.NewPod(1),
				test.NewPod(3),
				test.NewPod(4),
			},
		},
		{
			name:           "currentRevDiff less or equal than zero, choose updated pods to delete",
			totalDiff:      1,
			currentRevDiff: 0,
			notUpdatedPods: []*corev1.Pod{
				test.NewPod(1),
				test.NewPod(2),
			},
			updatedPods: []*corev1.Pod{
				test.NewPod(3),
				test.NewPod(4),
				test.NewPod(5),
			},
			expectedPods: []*corev1.Pod{
				test.NewPod(3),
			},
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			if pods := choosePodsToDelete(s.totalDiff, s.currentRevDiff, s.notUpdatedPods, s.updatedPods, s.sortMethod,
				nodeLister); !reflect.DeepEqual(s.expectedPods, pods) {
				t.Errorf("expected pods: %v, got pods: %v", s.expectedPods, pods)
			}
		})
	}
}

func TestGetCostFromPod(t *testing.T) {
	tests := []struct {
		name         string
		deletionCost string
		method       string
		expected     float64
	}{
		{
			name:         "max edgeCase",
			deletionCost: "",
			method:       "",
			expected:     math.MaxFloat64,
		},
		{
			name:         "min edgeCase",
			deletionCost: "",
			method:       CostSortMethodDescend,
			expected:     -math.MaxFloat64,
		},
		{
			name:         "wrong deletion cost",
			deletionCost: "ff",
			method:       "",
			expected:     math.MaxFloat64,
		},
		{
			name:         "correct deletion cost",
			deletionCost: "10",
			method:       "",
			expected:     10,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			pod := test.NewPod(1)
			pod.Annotations = map[string]string{
				PodDeletionCost: s.deletionCost,
			}
			if got := getCostFromPod(pod, s.method); got != s.expected {
				t.Errorf("expected %f, got %f", s.expected, got)
			}
		})
	}
}

func TestGenAvailableIndex(t *testing.T) {
	pod1 := test.NewPod(1)
	pod1.Annotations[v1alpha1.GameDeploymentIndexID] = "1"
	pod2 := test.NewPod(2)
	pod2.Annotations[v1alpha1.GameDeploymentIndexID] = "2"
	pod3 := test.NewPod(3)
	pod3.Annotations[v1alpha1.GameDeploymentIndexID] = "3"
	if indexs := genAvailableIndex(true, 1, 5, []*corev1.Pod{pod1, pod2, pod3}); !reflect.DeepEqual(indexs, []int{4}) {
		t.Errorf("expected [4], got %v", indexs)
	}
}

func TestSortPodsByAnnotations(t *testing.T) {
	nodeLister := newFakeNodes()
	pod1 := test.NewPod(1)
	pod1.Annotations[PodDeletionCost] = "1"
	pod1.Spec.NodeName = "node-cost-20"

	pod2 := test.NewPod(2)
	pod2.Annotations[PodDeletionCost] = "1"
	pod2.Spec.NodeName = "node-cost-10"

	pod3 := test.NewPod(3)
	pod3.Annotations[PodDeletionCost] = "3"
	pod3.Spec.NodeName = "node-cost-20"

	pod4 := test.NewPod(4)
	pod4.Annotations[PodDeletionCost] = "4"
	pod4.Spec.NodeName = "node-cost-10"

	pod5 := test.NewPod(5)
	pod5.Spec.NodeName = "node-cost-10"

	pod6 := test.NewPod(6)
	pod6.Annotations[PodDeletionCost] = "6"
	pod6.Spec.NodeName = "node-without-cost"

	pod7 := test.NewPod(7)
	pod7.Annotations[PodDeletionCost] = "7"
	pod7.Spec.NodeName = "node-cost-10"

	pod8 := test.NewPod(8)
	pod8.Annotations[PodDeletionCost] = "-8"
	pod8.Spec.NodeName = "node-cost-ff"

	type args struct {
		pods       []*corev1.Pod
		nodeLister corelisters.NodeLister
		sortMethod string
	}
	tests := []struct {
		name     string
		args     args
		expected []*corev1.Pod
	}{
		// TODO: Add test cases.
		{
			name: "different node cost with same pod cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod2},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod2, pod1},
		},
		{
			name: "same node cost with different pod cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod3},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod1, pod3},
		},
		{
			name: "different node cost with different pod cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod4},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod4, pod1},
		},
		{
			name: "different node cost with one null pod cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod5},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod5, pod1},
		},
		{
			name: "different pod cost with one null node cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod6},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod1, pod6},
		},
		{
			name: "same node cost with different pod cost, Descend",
			args: args{
				pods:       []*corev1.Pod{pod1, pod7},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodDescend,
			},
			expected: []*corev1.Pod{pod7, pod1},
		},
		{
			name: "different pod cost with one error node cost",
			args: args{
				pods:       []*corev1.Pod{pod1, pod8},
				nodeLister: nodeLister,
				sortMethod: CostSortMethodAscend,
			},
			expected: []*corev1.Pod{pod1, pod8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortPodsByAnnotations(tt.args.pods, tt.args.nodeLister, tt.args.sortMethod)
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("expected %+v, got %+v", tt.expected[i].Name, got[i].Name)
				}
			}
		})
	}
}
