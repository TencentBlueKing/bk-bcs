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

// Package cluster xxx
package cluster

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// ListFederationNamespaces list federation namespaces from host federation cluster
func (h *clusterClient) ListFederationNamespaces(hostClusterID string) ([]*FederationNamespace, error) {
	url := fmt.Sprintf("%s%s", h.opt.Endpoint, fmt.Sprintf(ListK8SNamespacePath, hostClusterID))
	blog.Infof("listing all Federation Namespaces for host cluster: %s, request url: %s", hostClusterID, url)

	raw, err := h.opt.Sender.DoGetRequest(url, h.defaultHeader)
	if err != nil {
		return nil, fmt.Errorf("ListFederationNamespaces failed when DoGetRequest error: %s",
			err.Error())
	}
	var k8sNamespaceList v1.NamespaceList
	if err := json.Unmarshal(raw, &k8sNamespaceList); err != nil {
		return nil, fmt.Errorf("decode NodeList response failed %s, raw response %s", err.Error(), string(raw))
	}

	fedNamespaceList := make([]*FederationNamespace, 0)
	if len(k8sNamespaceList.Items) == 0 {
		blog.Infof("namespace list is empty for cluster: %s, skip", hostClusterID)
		return fedNamespaceList, nil
	}

	for _, ns := range k8sNamespaceList.Items {
		// check is federated namespace
		isFedNamespace, ok := ns.Annotations[FedNamespaceIsFederatedKey]
		if !ok || isFedNamespace != "true" {
			continue
		}

		//todo 此时如果联邦命名空间没有指定作用的集群范围，则应该默认为所有集群，此处没有把全部子集群传入都作为默认集群
		// cluster range. e.g. bcs-k8s-00001,bcs-k8s-00002,bcs-k8s-00003
		clusterRangeStr := ns.Annotations[FedNamespaceClusterRangeKey]
		subClusterIds := make([]string, 0)
		if len(clusterRangeStr) != 0 {
			lower := strings.Split(clusterRangeStr, ",")
			for _, sc := range lower {
				subClusterIds = append(subClusterIds, strings.ToUpper(sc))
			}
		}
		// project code. e.g. default
		projectCode := ns.Annotations[FedNamespaceProjectCodeKey]

		fedNamespace := &FederationNamespace{
			HostClusterId: hostClusterID,
			Namespace:     ns.Name,
			SubClusters:   subClusterIds,
			ProjectCode:   projectCode,
			CreatedTime:   ns.CreationTimestamp.Time,
		}
		fedNamespaceList = append(fedNamespaceList, fedNamespace)
	}

	return fedNamespaceList, nil
}
