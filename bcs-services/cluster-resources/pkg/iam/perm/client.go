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

package perm

import (
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// IAMCacheTTL IAM 权限校验缓存时间，单位：s
const IAMCacheTTL = 10

// IAMClient xxx
type IAMClient struct {
	cli func(tenantID string) iam.PermClient
}

// NewIAMClient xxx
func NewIAMClient() *IAMClient {
	return &IAMClient{cli: conf.G.IAM.Cli}
}

func (c *IAMClient) permReq(username string) iam.PermissionRequest {
	return iam.PermissionRequest{
		SystemID: conf.G.IAM.SystemID,
		UserName: username,
	}
}

// ResTypeAllowed 判断用户是否具备某个操作权限（资源实例无关）
func (c *IAMClient) ResTypeAllowed(tenantID, username, actionID string, useCache bool) (bool, error) {
	return c.cli(tenantID).IsAllowedWithoutResource(actionID, c.permReq(username), useCache)
}

// ResInstAllowed 判断用户对某个资源实例是否具有指定操作的权限
func (c *IAMClient) ResInstAllowed(
	tenantID, username, actionID string, resources []bkiam.ResourceNode, useCache bool,
) (bool, error) {
	return c.cli(tenantID).IsAllowedWithResource(actionID, c.permReq(username), convertResourceNodes(resources), useCache)
}

// ResTypeMultiActionsAllowed 判断用户是否具备多个操作的权限
func (c *IAMClient) ResTypeMultiActionsAllowed(tenantID, username string, actionIDs []string) (map[string]bool, error) {
	ret := map[string]bool{}
	for _, id := range actionIDs {
		allow, err := c.ResTypeAllowed(tenantID, username, id, false)
		if err != nil {
			return ret, err
		}
		ret[id] = allow
	}
	return ret, nil
}

// ResInstMultiActionsAllowed 判断用户对某个(单个)资源实例是否具有多个操作的权限.
func (c *IAMClient) ResInstMultiActionsAllowed(
	tenantID, username string, actionIDs []string, resources []bkiam.ResourceNode,
) (map[string]bool, error) {
	return c.cli(tenantID).ResourceMultiActionsAllowed(actionIDs, c.permReq(username), convertResourceNodes(resources))
}

// BatchResMultiActionsAllowed 判断用户对某些资源是否具有多个指定操作的权限. 当前sdk仅支持同类型的资源
func (c *IAMClient) BatchResMultiActionsAllowed(
	tenantID, username string, actionsIDs []string, resources []bkiam.ResourceNode,
) (map[string]map[string]bool, error) {
	nodes := make([][]iam.ResourceNode, 0, len(resources))
	for _, res := range resources {
		nodes = append(nodes, []iam.ResourceNode{convertResourceNode(res)})
	}
	return c.cli(tenantID).BatchResourceMultiActionsAllowed(actionsIDs, c.permReq(username), nodes)
}
