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

// GetTask 获取任务详情
func GetTask(ctx context.Context, req *clustermanager.GetTaskRequest) (*clustermanager.Task, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.GetTask(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetTask error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("GetTask error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// RetryTask 任务重试
func RetryTask(ctx context.Context, req *clustermanager.RetryTaskRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.RetryTask(ctx, req)
	if err != nil {
		return false, fmt.Errorf("RetryTask error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("RetryTask error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}

// SkipTask 跳过当前任务
func SkipTask(ctx context.Context, req *clustermanager.SkipTaskRequest) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.SkipTask(ctx, req)
	if err != nil {
		return false, fmt.Errorf("SkipTask error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("SkipTask error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Result, nil
}
