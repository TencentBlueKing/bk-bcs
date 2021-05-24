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

package listenerclient

import (
	"context"
	"fmt"

	cloudListenerType "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/network/v1"
	listenerClientV1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/clientset/versioned/typed/network/v1"
	listerV1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/listers/network/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type ListenerClient struct {
	client  listenerClientV1.NetworkV1Interface
	lister  listerV1.CloudListenerLister
	clbName string
}

func NewListenerClient(clbname string, client listenerClientV1.NetworkV1Interface, lister listerV1.CloudListenerLister) (Interface, error) {

	return &ListenerClient{
		client:  client,
		lister:  lister,
		clbName: clbname,
	}, nil
}

func (lc *ListenerClient) ListListeners() ([]*cloudListenerType.CloudListener, error) {
	selector := labels.NewSelector()
	requirement, err := labels.NewRequirement("bmsf.tencent.com/clbname", selection.Equals, []string{lc.clbName})
	if err != nil {
		return nil, fmt.Errorf("create requirement failed, err %s", err.Error())
	}
	selector = selector.Add(*requirement)
	return lc.lister.List(selector)
}
func (lc *ListenerClient) Create(listener *cloudListenerType.CloudListener) error {
	_, err := lc.client.CloudListeners(listener.GetNamespace()).Create(
		context.Background(), listener, metav1.CreateOptions{})
	return err
}
func (lc *ListenerClient) Update(listener *cloudListenerType.CloudListener) error {
	old, err := lc.lister.CloudListeners(listener.GetNamespace()).Get(listener.GetName())
	if err != nil {
		_, createErr := lc.client.CloudListeners(listener.GetNamespace()).Create(
			context.Background(), listener, metav1.CreateOptions{})
		return createErr
	}
	listener.SetResourceVersion(old.GetResourceVersion())
	_, err = lc.client.CloudListeners(listener.GetNamespace()).Update(
		context.Background(), listener, metav1.UpdateOptions{})
	return err
}
func (lc *ListenerClient) Delete(listener *cloudListenerType.CloudListener) error {
	return lc.client.CloudListeners(listener.GetNamespace()).Delete(
		context.Background(), listener.GetName(), metav1.DeleteOptions{})
}
