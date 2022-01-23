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
	"fmt"
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/testutil"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/controller"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestStorageMatches(t *testing.T) {
	tests := []struct {
		name             string
		setName          string
		volumeClaimNames []string
		podName          string
		podVolumes       []corev1.Volume
		expected         bool
	}{
		{
			name:             "matched",
			setName:          "test-set",
			volumeClaimNames: []string{"data", "log"},
			podName:          "test-set-0",
			podVolumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "data-test-set-0",
						},
					},
				},
				{
					Name: "log",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "log-test-set-0",
						},
					},
				},
			},
			expected: true,
		},
		{
			name:             "not match",
			setName:          "test-set",
			volumeClaimNames: []string{"data", "log"},
			podName:          "test-set-0",
			podVolumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "data-test-set-0",
						},
					},
				},
				{
					Name: "backup",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "backup-test-set-0",
						},
					},
				},
			},
			expected: false,
		},
		{
			name:             "not found",
			setName:          "test-set",
			volumeClaimNames: []string{"data", "log"},
			podName:          "test-set-0",
			podVolumes:       []corev1.Volume{},
			expected:         false,
		},
		{
			name:     "test invalid pod",
			podName:  "test-set-f",
			expected: false,
		},
	}

	for _, s := range tests {
		t.Run(s.name, func(t *testing.T) {
			set := testutil.NewGameStatefulSet(1)
			set.Name = s.setName
			for _, name := range s.volumeClaimNames {
				set.Spec.VolumeClaimTemplates = append(set.Spec.VolumeClaimTemplates, corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name: name,
					},
				})
			}
			pod := testutil.NewPod(0)
			pod.Name = s.podName
			pod.Spec.Volumes = s.podVolumes
			if got := storageMatches(set, pod); got != s.expected {
				t.Errorf("storageMatches() = %v, want %v", got, s.expected)
			}
		})
	}
}

func TestNewVersionedGameStatefulSetPod(t *testing.T) {
	set := testutil.NewGameStatefulSet(1)
	set.Spec.UpdateStrategy.Type = gstsv1alpha1.InplaceUpdateGameStatefulSetStrategyType
	set.Spec.UpdateStrategy.CanaryStrategy = &gstsv1alpha1.CanaryStrategy{
		Steps: []gstsv1alpha1.CanaryStep{{Pause: nil}},
	}
	currentSet := testutil.NewGameStatefulSet(1)
	updateSet := testutil.NewGameStatefulSet(1)
	pod := newStatefulSetPod(set, 0)
	pod.Labels["controller-revision-hash"] = "1"
	pod.Spec.ReadinessGates = append(pod.Spec.ReadinessGates, corev1.PodReadinessGate{ConditionType: "InPlaceUpdateReady"})

	if got := newVersionedGameStatefulSetPod(set, currentSet, updateSet, "1", "2", 0); !reflect.DeepEqual(got, pod) {
		t.Errorf("newVersionedGameStatefulSetPod error, \ngot: \t%v\nwant: \t%v", got, pod)
	}
}

func newStatefulSet(replicas int, label bool) *gstsv1alpha1.GameStatefulSet {
	petMounts := []corev1.VolumeMount{
		{Name: "data", MountPath: "/data"},
	}
	podMounts := []corev1.VolumeMount{
		{Name: "log", MountPath: "/var/log"},
	}
	var labels = map[string]string{}
	if label {
		labels = map[string]string{"foo": "bar"}
	} else {
		labels = nil
	}
	return newStatefulSetWithVolumes(replicas, "foo", petMounts, podMounts, labels)
}

// newStatefulSetPod returns a new Pod conforming to the set's Spec with an identity generated from ordinal.
func newStatefulSetPod(set *gstsv1alpha1.GameStatefulSet, ordinal int) *corev1.Pod {
	pod, _ := controller.GetPodFromTemplate(&set.Spec.Template, set, metav1.NewControllerRef(set,
		gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet")))
	pod.Name = getPodName(set, ordinal)
	initIdentity(set, pod)
	updateStorage(set, pod)
	return pod
}

func newPVC(name string, labels map[string]string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(1, resource.BinarySI),
				},
			},
		},
	}
}

func newStatefulSetWithVolumes(replicas int, name string, petMounts []corev1.VolumeMount, podMounts []corev1.VolumeMount, labels map[string]string) *gstsv1alpha1.GameStatefulSet {
	mounts := append(petMounts, podMounts...)
	claims := []corev1.PersistentVolumeClaim{}
	for _, m := range petMounts {
		claims = append(claims, *newPVC(m.Name, labels))
	}

	vols := []corev1.Volume{}
	for _, m := range podMounts {
		vols = append(vols, corev1.Volume{
			Name: m.Name,
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("/tmp/%v", m.Name),
				},
			},
		})
	}

	template := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:         "nginx",
					Image:        "nginx",
					VolumeMounts: mounts,
				},
			},
			Volumes: vols,
		},
	}

	template.Labels = map[string]string{"foo": "bar"}

	return &gstsv1alpha1.GameStatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GameStatefulSet",
			APIVersion: "tkex.tencent.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
			UID:       types.UID("test"),
		},
		Spec: gstsv1alpha1.GameStatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Replicas:             func() *int32 { i := int32(replicas); return &i }(),
			Template:             template,
			VolumeClaimTemplates: claims,
			ServiceName:          "governingsvc",
			UpdateStrategy:       gstsv1alpha1.GameStatefulSetUpdateStrategy{Type: gstsv1alpha1.RollingUpdateGameStatefulSetStrategyType},
			RevisionHistoryLimit: func() *int32 {
				limit := int32(2)
				return &limit
			}(),
		},
	}
}

func TestUpdateStorage(t *testing.T) {
	set := newStatefulSet(3, false)
	pod := newStatefulSetPod(set, 1)
	if !storageMatches(set, pod) {
		t.Error("Newly created Pod has a invalid storage")
	}
	pod.Spec.Volumes = nil
	if storageMatches(set, pod) {
		t.Error("Pod with invalid Volumes has valid storage")
	}
	updateStorage(set, pod)
	if !storageMatches(set, pod) {
		t.Error("updateStorage failed to recreate volumes")
	}
	pod = newStatefulSetPod(set, 1)
	for i := range pod.Spec.Volumes {
		pod.Spec.Volumes[i].PersistentVolumeClaim = nil
	}
	if storageMatches(set, pod) {
		t.Error("Pod with invalid Volumes claim valid storage")
	}
	updateStorage(set, pod)
	if !storageMatches(set, pod) {
		t.Error("updateStorage failed to recreate volume claims")
	}
	pod = newStatefulSetPod(set, 1)
	for i := range pod.Spec.Volumes {
		if pod.Spec.Volumes[i].PersistentVolumeClaim != nil {
			pod.Spec.Volumes[i].PersistentVolumeClaim.ClaimName = "foo"
		}
	}
	if storageMatches(set, pod) {
		t.Error("Pod with invalid Volumes claim valid storage")
	}
	updateStorage(set, pod)
	if !storageMatches(set, pod) {
		t.Error("updateStorage failed to recreate volume claim names")
	}
	set1 := newStatefulSet(3, true)
	pod1 := newStatefulSetPod(set1, 1)
	updateStorage(set1, pod1)
	if !storageMatches(set1, pod1) {
		t.Error("updateStorage failed when pvc with labels")
	}
}

func newGameStatefulSetWithTime(name string, t time.Time) *gstsv1alpha1.GameStatefulSet {
	set := testutil.NewGameStatefulSet(1)
	set.Name = name
	set.CreationTimestamp = metav1.NewTime(t)
	return set
}

func TestOverlappingGameStatefulSetes(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(1 * time.Second)
	set1 := newGameStatefulSetWithTime("a", t2)
	set2 := newGameStatefulSetWithTime("c", t2)
	set3 := newGameStatefulSetWithTime("b", t2)
	set4 := newGameStatefulSetWithTime("f", t1)

	sets := []*gstsv1alpha1.GameStatefulSet{set1, set2, set3, set4}
	sort.Sort(overlappingGameStatefulSetes(sets))

	wantNames := []string{"f", "a", "b", "c"}
	var got []string
	for _, set := range sets {
		got = append(got, set.Name)
	}
	if !reflect.DeepEqual(got, wantNames) {
		t.Errorf("Overlapping GameStatefulSets returned in unexpected order. Got: %v, want: %v", got, wantNames)
	}
}

func TestInconsistentStatus(t *testing.T) {
	sts := testutil.NewGameStatefulSet(1)
	sts.Status = gstsv1alpha1.GameStatefulSetStatus{
		ObservedGeneration: 1,
		Replicas:           1,
		CurrentReplicas:    1,
		ReadyReplicas:      2,
		UpdatedReplicas:    1,
		CurrentRevision:    "1",
		UpdateRevision:     "1",
	}

	newStatus := gstsv1alpha1.GameStatefulSetStatus{
		ObservedGeneration: 1,
		Replicas:           1,
		CurrentReplicas:    1,
		ReadyReplicas:      2,
		UpdatedReplicas:    1,
		CurrentRevision:    "1",
		UpdateRevision:     "2",
	}

	if !inconsistentStatus(sts, &newStatus) {
		t.Errorf("consistent status")
	}
}
