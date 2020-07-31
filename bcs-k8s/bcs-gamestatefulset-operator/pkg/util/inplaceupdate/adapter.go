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

package inplaceupdate

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

type adapter interface {
	getPod(namespace, name string) (*v1.Pod, error)
	updatePod(pod *v1.Pod) error
	updatePodStatus(pod *v1.Pod) error
}

type adapterTypedClient struct {
	client clientset.Interface
}

func (c *adapterTypedClient) getPod(namespace, name string) (*v1.Pod, error) {
	return c.client.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
}

func (c *adapterTypedClient) updatePod(pod *v1.Pod) error {
	_, err := c.client.CoreV1().Pods(pod.Namespace).Update(pod)
	return err
}

func (c *adapterTypedClient) updatePodStatus(pod *v1.Pod) error {
	_, err := c.client.CoreV1().Pods(pod.Namespace).UpdateStatus(pod)
	return err
}
