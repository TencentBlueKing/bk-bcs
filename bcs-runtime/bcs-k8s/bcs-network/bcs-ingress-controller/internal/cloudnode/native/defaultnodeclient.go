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

package native

import (
	"errors"

	corev1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// NativeNodeClient default node exporter
type NativeNodeClient struct {
}

// NewNativeNodeClient return new native node client
func NewNativeNodeClient() *NativeNodeClient {
	return &NativeNodeClient{}
}

// GetNodeExternalIpList 获取node.Status.Addresses下的ExternalIP
func (n *NativeNodeClient) GetNodeExternalIpList(node *corev1.Node) ([]string, error) {
	externalIpList := make([]string, 0)

	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			externalIpList = append(externalIpList, addr.Address)
		}
	}

	if len(externalIpList) == 0 {
		return nil, errors.New("empty node external ip list")
	}

	blog.Infof("get node %s ip list: %v", node.Name, externalIpList)
	return externalIpList, nil
}
