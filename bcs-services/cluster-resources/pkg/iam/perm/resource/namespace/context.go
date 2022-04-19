/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package namespace

import (
	"strings"

	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/hash"
)

// PermCtx ...
type PermCtx struct {
	Username    string
	ProjectID   string
	ClusterID   string
	Namespace   string
	NamespaceID string
	forceRaise  bool
}

// NewPermCtx ...
func NewPermCtx(username, projectID, clusterID, namespace string) *PermCtx {
	return &PermCtx{
		Username:    username,
		ProjectID:   projectID,
		ClusterID:   clusterID,
		Namespace:   namespace,
		NamespaceID: calcNamespaceID(clusterID, namespace),
	}
}

// Validate ...
func (c *PermCtx) Validate(actionIDs []string) error {
	if c.Username == "" || c.ProjectID == "" || c.ClusterID == "" {
		return errorx.New(errcode.ValidateErr, "Ctx validate failed: Username/ProjectID/ClusterID required")
	}
	// 如果是 命名空间创建，获取列表 可不需要命名空间
	for _, actionID := range actionIDs {
		if actionID == NamespaceList || actionID == NamespaceCreate {
			continue
		}
		if c.Namespace == "" {
			return errorx.New(errcode.ValidateErr, "Ctx validate failed: Namespace required")
		}
	}
	return nil
}

// GetProjID ...
func (c *PermCtx) GetProjID() string {
	return c.ProjectID
}

// GetClusterID ...
func (c *PermCtx) GetClusterID() string {
	return c.ClusterID
}

// GetResID ...
func (c *PermCtx) GetResID() string {
	return c.NamespaceID
}

// GetUsername ...
func (c *PermCtx) GetUsername() string {
	return c.Username
}

// GetParentChain ...
func (c *PermCtx) GetParentChain() []perm.IAMRes {
	return []perm.IAMRes{
		{ResType: perm.ResTypeProj, ResID: c.ProjectID},
		{ResType: perm.ResTypeCluster, ResID: c.ClusterID},
	}
}

// SetForceRaise ...
func (c *PermCtx) SetForceRaise() {
	c.forceRaise = true
}

// ForceRaise ...
func (c *PermCtx) ForceRaise() bool {
	return c.forceRaise
}

// ToMap ...
func (c *PermCtx) ToMap() map[string]interface{} {
	return structs.Map(c)
}

// FromMap ...
func (c *PermCtx) FromMap(m map[string]interface{}) perm.Ctx {
	if username, ok := m["Username"]; ok {
		c.Username = username.(string)
	}
	if projID, ok := m["ProjectID"]; ok {
		c.ProjectID = projID.(string)
	}
	if clusterID, ok := m["ClusterID"]; ok {
		c.ClusterID = clusterID.(string)
		// ClusterID, Namespace 只要有变动，都重新计算 NamespaceID
		c.NamespaceID = calcNamespaceID(c.ClusterID, c.Namespace)
	}
	if namespace, ok := m["Namespace"]; ok {
		c.Namespace = namespace.(string)
		c.NamespaceID = calcNamespaceID(c.ClusterID, c.Namespace)
	}
	return c
}

// 计算(压缩)出注册到权限中心的命名空间 ID，具备唯一性. 当前的算法并不能完全避免冲突，但概率较低。
// note: 权限中心对资源 ID 有长度限制，不超过32位。长度越长，处理性能越低
// NamespaceID 是命名空间注册到权限中心的资源 ID，它是对结构`集群ID:命名空间name`的一个压缩，
// 如 `BCS-K8S-40000:default` 会被处理成 `40000:5f03d33dde`。
func calcNamespaceID(clusterID, namespace string) string {
	if clusterID == "" || namespace == "" {
		return ""
	}
	clusterIDx := clusterID[strings.LastIndex(clusterID, "-")+1:]
	return clusterIDx + ":" + hash.MD5Digest(namespace)[8:16] + namespace[:2]
}
