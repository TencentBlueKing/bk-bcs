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

package namespace

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm/resource/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/basic"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/hash"
)

// calcNamespaceID 计算(压缩)出注册到权限中心的命名空间 ID，具备唯一性. 当前的算法并不能完全避免冲突，但概率较低。
// note: 权限中心对资源 ID 有长度限制，不超过32位。长度越长，处理性能越低
// NamespaceID 是命名空间注册到权限中心的资源 ID，它是对结构`集群ID:命名空间name`的一个压缩，
// 如 `BCS-K8S-40000:default` 会被处理成 `40000:5f03d33dde`。
func calcNamespaceID(clusterID, namespace string) string {
	if clusterID == "" || namespace == "" {
		return ""
	}
	clusterIDx := clusterID[strings.LastIndex(clusterID, "-")+1:]
	return clusterIDx + ":" + hash.MD5Digest(namespace)[8:16] + namespace[:basic.MinInt(2, len(namespace))]
}

// FetchBatchNSScopedResMultiActPerm 获取对多个命名空间域资源的 CURD 权限信息
func FetchBatchNSScopedResMultiActPerm(
	username, projectID, clusterID string, namespaces []string, res string,
) (map[string]map[string]bool, error) {
	// 命名空间域资源 CURD 操作
	actionIDs := []string{NamespaceScopedView, NamespaceScopedCreate, NamespaceScopedUpdate, NamespaceScopedDelete}

	// 转换成命名空间 ID，记录对应关系
	id2nsMap := map[string]string{}
	resIDs := []string{}
	for _, ns := range namespaces {
		nsID := calcNamespaceID(clusterID, ns)
		resIDs = append(resIDs, nsID)
		id2nsMap[nsID] = ns
	}

	iamPermCli := perm.IAMPerm{
		Cli:           perm.NewIAMClient(),
		ResType:       perm.ResTypeNS,
		PermCtx:       &PermCtx{},
		ResReq:        NewResRequest(projectID, clusterID, res),
		ParentResPerm: &cluster.NewPerm(projectID).IAMPerm,
	}
	ret, err := iamPermCli.BatchResMultiActionAllowed(username, actionIDs, resIDs)
	if err != nil {
		return nil, err
	}

	// 转换回命名空间：权限信息
	nsScopePerms := map[string]map[string]bool{}
	for nsID, scopePerm := range ret {
		nsScopePerms[id2nsMap[nsID]] = scopePerm
	}
	return nsScopePerms, nil
}
