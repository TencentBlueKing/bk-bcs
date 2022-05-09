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

package perm

import (
	bkiam "github.com/TencentBlueKing/iam-go-sdk"

	conf "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
)

// IAMCacheTTL IAM 权限校验缓存时间，单位：s
const IAMCacheTTL = 10

// IAMClient ...
type IAMClient struct {
	cli *bkiam.IAM
}

// NewIAMClient ...
func NewIAMClient() *IAMClient {
	return &IAMClient{cli: conf.G.IAM.Cli}
}

// ResTypeAllowed 判断用户是否具备某个操作权限（资源实例无关）
func (c *IAMClient) ResTypeAllowed(username, actionID string, useCache bool) (bool, error) {
	req := c.makeRequest(username, actionID, []bkiam.ResourceNode{})
	if useCache {
		return c.cli.IsAllowedWithCache(req, IAMCacheTTL)
	}
	return c.cli.IsAllowed(req)
}

// ResInstAllowed 判断用户对某个资源实例是否具有指定操作的权限
func (c *IAMClient) ResInstAllowed(username, actionID string, resources []bkiam.ResourceNode, useCache bool) (bool, error) {
	req := c.makeRequest(username, actionID, resources)
	if useCache {
		return c.cli.IsAllowedWithCache(req, IAMCacheTTL)
	}
	return c.cli.IsAllowed(req)
}

// ResTypeMultiActionsAllowed 判断用户是否具备多个操作的权限
func (c *IAMClient) ResTypeMultiActionsAllowed(username string, actionIDs []string) (map[string]bool, error) {
	ret := map[string]bool{}
	for _, id := range actionIDs {
		allow, err := c.ResTypeAllowed(username, id, false)
		if err != nil {
			return ret, err
		}
		ret[id] = allow
	}
	return ret, nil
}

// ResInstMultiActionsAllowed 判断用户对某个(单个)资源实例是否具有多个操作的权限.
func (c *IAMClient) ResInstMultiActionsAllowed(
	username string, actionIDs []string, resources []bkiam.ResourceNode,
) (map[string]bool, error) {
	req := c.makeMultiActionRequest(username, actionIDs, resources)
	return c.cli.ResourceMultiActionsAllowed(req)
}

// BatchResMultiActionsAllowed 判断用户对某些资源是否具有多个指定操作的权限. 当前sdk仅支持同类型的资源
func (c *IAMClient) BatchResMultiActionsAllowed(
	username string, actionsIDs []string, resources []bkiam.ResourceNode,
) (map[string]map[string]bool, error) {
	req := c.makeMultiActionRequest(username, actionsIDs, []bkiam.ResourceNode{})
	resourceList := []bkiam.Resources{}
	for _, res := range resources {
		resourceList = append(resourceList, []bkiam.ResourceNode{res})
	}
	return c.cli.BatchResourceMultiActionsAllowed(req, resourceList)
}

func (c *IAMClient) makeRequest(username, actionID string, resources []bkiam.ResourceNode) bkiam.Request {
	return bkiam.Request{
		System:    conf.G.IAM.SystemID,
		Subject:   bkiam.Subject{Type: "user", ID: username},
		Action:    bkiam.Action{ID: actionID},
		Resources: resources,
	}
}

func (c *IAMClient) makeMultiActionRequest(
	username string, actionIDs []string, resources []bkiam.ResourceNode,
) bkiam.MultiActionRequest {
	actions := []bkiam.Action{}
	for _, id := range actionIDs {
		actions = append(actions, bkiam.Action{ID: id})
	}
	return bkiam.MultiActionRequest{
		System:    conf.G.IAM.SystemID,
		Subject:   bkiam.Subject{Type: "user", ID: username},
		Actions:   actions,
		Resources: resources,
	}
}
