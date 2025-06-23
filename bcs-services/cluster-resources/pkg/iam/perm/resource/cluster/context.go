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

package cluster

import (
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// PermCtx xxx
type PermCtx struct {
	Username   string
	ProjectID  string
	ClusterID  string
	TenantID   string
	forceRaise bool
}

// NewPermCtx xxx
func NewPermCtx(username, projectID, clusterID, tenantID string) *PermCtx {
	return &PermCtx{
		Username:  username,
		ProjectID: projectID,
		ClusterID: clusterID,
		TenantID:  tenantID,
	}
}

// Validate xxx
func (c *PermCtx) Validate(_ []string) error {
	if c.Username == "" || c.ProjectID == "" || c.ClusterID == "" {
		return errorx.New(errcode.ValidateErr, "Ctx validate failed: Username/ProjectID/ClusterID required")
	}
	return nil
}

// GetProjID xxx
func (c *PermCtx) GetProjID() string {
	return c.ProjectID
}

// GetClusterID xxx
func (c *PermCtx) GetClusterID() string {
	return c.ClusterID
}

// GetResID xxx
func (c *PermCtx) GetResID() string {
	return c.ClusterID
}

// GetUsername xxx
func (c *PermCtx) GetUsername() string {
	return c.Username
}

// GetTenantID xxx
func (c *PermCtx) GetTenantID() string {
	return c.TenantID
}

// GetParentChain xxx
func (c *PermCtx) GetParentChain() []perm.IAMRes {
	return []perm.IAMRes{
		{ResType: perm.ResTypeProj, ResID: c.ProjectID},
	}
}

// SetForceRaise xxx
func (c *PermCtx) SetForceRaise() {
	c.forceRaise = true
}

// ForceRaise xxx
func (c *PermCtx) ForceRaise() bool {
	return c.forceRaise
}

// ToMap xxx
func (c *PermCtx) ToMap() map[string]interface{} {
	return structs.Map(c)
}

// FromMap xxx
func (c *PermCtx) FromMap(m map[string]interface{}) perm.Ctx {
	if username, ok := m["Username"]; ok {
		c.Username = username.(string)
	}
	if projID, ok := m["ProjectID"]; ok {
		c.ProjectID = projID.(string)
	}
	if clusterID, ok := m["ClusterID"]; ok {
		c.ClusterID = clusterID.(string)
	}
	if tenantID, ok := m["TenantID"]; ok {
		c.TenantID = tenantID.(string)
	}
	return c
}

// GetNamespace xxx
func (c *PermCtx) GetNamespace() string {
	return ""
}
