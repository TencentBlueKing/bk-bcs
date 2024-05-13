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

package check

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func TestGetPortConflictMap(t *testing.T) {
	pChecker := &PortBindChecker{}
	portbindingList := &networkextensionv1.PortBindingList{
		Items: []networkextensionv1.PortBinding{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-1",
					Namespace: "default",
				},
				Spec: networkextensionv1.PortBindingSpec{
					PortBindingList: []*networkextensionv1.PortBindingItem{
						{
							PoolName:      "pool1",
							PoolNamespace: "default",
							PoolItemName:  "item1",
							StartPort:     1230,
							EndPort:       1250,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-2",
					Namespace: "default",
				},
				Spec: networkextensionv1.PortBindingSpec{
					PortBindingList: []*networkextensionv1.PortBindingItem{
						{
							PoolName:      "pool1",
							PoolNamespace: "default",
							PoolItemName:  "item1",
							StartPort:     1250,
							EndPort:       1270,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-3",
					Namespace: "default",
				},
				Spec: networkextensionv1.PortBindingSpec{
					PortBindingList: []*networkextensionv1.PortBindingItem{
						{
							PoolName:      "pool1",
							PoolNamespace: "default",
							PoolItemName:  "item1",
							StartPort:     1230,
							EndPort:       1250,
						},
					},
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod-4",
					Namespace: "default",
				},
				Spec: networkextensionv1.PortBindingSpec{
					PortBindingList: []*networkextensionv1.PortBindingItem{
						{
							PoolName:      "pool1",
							PoolNamespace: "default",
							PoolItemName:  "item1",
							StartPort:     1250,
							EndPort:       1270,
						},
					},
				},
			},
		},
	}
	conflictMap := pChecker.getPortConflictMap(portbindingList)
	fmt.Println(conflictMap)
}
