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

package cluster

import (
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/iam/perm"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// PermCtx ...
type PermCtx struct {
	Username   string
	ProjectID  string
	ClusterID  string
	forceRaise bool
}

// NewPermCtx ...
func NewPermCtx(username, projectID, clusterID string) *PermCtx {
	return &PermCtx{
		Username:  username,
		ProjectID: projectID,
		ClusterID: clusterID,
	}
}

// Validate ...
func (c *PermCtx) Validate(_ []string) error {
	if c.Username == "" || c.ProjectID == "" || c.ClusterID == "" {
		return errorx.New(errcode.ValidateErr, "Ctx validate failed: Username/ProjectID/ClusterID required")
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
	return c.ClusterID
}

// GetUsername ...
func (c *PermCtx) GetUsername() string {
	return c.Username
}

// GetParentChain ...
func (c *PermCtx) GetParentChain() []perm.IAMRes {
	return []perm.IAMRes{
		{ResType: perm.ResTypeProj, ResID: c.ProjectID},
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
	}
	return c
}
