package aggregation

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PodAggregation struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	Spec   corev1.PodSpec
	Status corev1.PodStatus
}

type PodAggregationList struct {
	metav1.TypeMeta
	metav1.ListMeta
	Items []PodAggregation
}