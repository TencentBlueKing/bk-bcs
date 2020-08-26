package update

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type Adapter interface {
	GetPod(namespace, name string) (*v1.Pod, error)
	UpdatePod(pod *v1.Pod) error
	UpdatePodStatus(pod *v1.Pod) error
}

type AdapterTypedClient struct {
	Client clientset.Interface
}

func (c *AdapterTypedClient) GetPod(namespace, name string) (*v1.Pod, error) {
	return c.Client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func (c *AdapterTypedClient) UpdatePod(pod *v1.Pod) error {
	_, err := c.Client.CoreV1().Pods(pod.Namespace).Update(pod)
	return err
}

func (c *AdapterTypedClient) UpdatePodStatus(pod *v1.Pod) error {
	_, err := c.Client.CoreV1().Pods(pod.Namespace).UpdateStatus(pod)
	return err
}
