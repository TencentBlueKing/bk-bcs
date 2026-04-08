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

// Package clustermanager cloudvpc操作
package clustermanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// ListNodeGroup 获取节点池列表
/*func ListNodeGroup(ctx context.Context, req *clustermanager.ListNodeGroupV2Request) (
	*clustermanager.ListNodeGroupResponseData, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.ListNodeGroupV2(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListNodeGroup error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("ListNodeGroup error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}*/

// GetNodeGroup 获取节点池信息
func GetNodeGroup(ctx context.Context, req *clustermanager.GetNodeGroupRequest) (
	*clustermanager.NodeGroup, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.GetNodeGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetNodeGroup error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("GetNodeGroup error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// EnableNodeGroupAutoScale 开启节点池节点自动扩缩容
func EnableNodeGroupAutoScale(ctx context.Context, req *clustermanager.EnableNodeGroupAutoScaleRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.EnableNodeGroupAutoScale(ctx, req)
	if err != nil {
		return false, fmt.Errorf("EnableNodeGroupAutoScale error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("EnableNodeGroupAutoScale error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}

// DisableNodeGroupAutoScale 关闭节点池节点自动扩缩容
func DisableNodeGroupAutoScale(ctx context.Context,
	req *clustermanager.DisableNodeGroupAutoScaleRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.DisableNodeGroupAutoScale(ctx, req)
	if err != nil {
		return false, fmt.Errorf("DisableNodeGroupAutoScale error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("DisableNodeGroupAutoScale error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}

// UpdateNodeGroup 更新节点池信息
func UpdateNodeGroup(ctx context.Context, req *clustermanager.UpdateNodeGroupRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.UpdateNodeGroup(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateNodeGroup error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("UpdateNodeGroup error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}

// UpdateGroupMinMaxSize 更新节点池minSize/maxSize信息
func UpdateGroupMinMaxSize(ctx context.Context, req *clustermanager.UpdateGroupMinMaxSizeRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.UpdateGroupMinMaxSize(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateGroupMinMaxSize error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("UpdateGroupMinMaxSize error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}
