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

package update

import (
	"context"
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
	return c.Client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c *AdapterTypedClient) UpdatePod(pod *v1.Pod) error {
	_, err := c.Client.CoreV1().Pods(pod.Namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}

func (c *AdapterTypedClient) UpdatePodStatus(pod *v1.Pod) error {
	_, err := c.Client.CoreV1().Pods(pod.Namespace).UpdateStatus(context.TODO(), pod, metav1.UpdateOptions{})
	return err
}
