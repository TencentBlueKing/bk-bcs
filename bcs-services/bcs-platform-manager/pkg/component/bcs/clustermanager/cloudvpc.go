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

// CreateCloudVPC 创建cloud vpc
func CreateCloudVPC(ctx context.Context, req *clustermanager.CreateCloudVPCRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return false, err
	}

	p, err := cli.CreateCloudVPC(ctx, req)
	if err != nil {
		return false, fmt.Errorf("CreateCloudVPC error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("CreateCloudVPC error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}

// UpdateCloudVPC 更新cloud vpc
func UpdateCloudVPC(ctx context.Context, req *clustermanager.UpdateCloudVPCRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	defer func() {
		if close != nil {
			close()
		}
	}()
	if err != nil {
		return false, err
	}

	p, err := cli.UpdateCloudVPC(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateCloudVPC error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("UpdateCloudVPC error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}
