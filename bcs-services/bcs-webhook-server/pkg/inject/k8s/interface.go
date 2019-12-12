package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

type K8sInject interface {
	InjectContent(*corev1.Pod) ([]PatchOperation, error)
}

type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
