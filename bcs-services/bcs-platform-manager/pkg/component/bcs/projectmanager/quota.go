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

// Package projectmanager xxx
package projectmanager

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

// CreateProjectQuota 申请项目资源额度
func CreateProjectQuota(ctx context.Context, req *bcsproject.CreateProjectQuotaRequest) (bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Quota.CreateProjectQuota(ctx, req)
	if err != nil {
		return false, fmt.Errorf("CreateProjectQuota error: %s", err)
	}

	if p.Code != 0 || p.Data == nil || p.Data.QuotaId == "" {
		return false, fmt.Errorf("CreateProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}

// GetProjectQuota 获取项目资源额度
func GetProjectQuota(ctx context.Context, quotaId string) (*bcsproject.ProjectQuotaResponse, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Quota.GetProjectQuota(ctx, &bcsproject.GetProjectQuotaRequest{QuotaId: quotaId})
	if err != nil {
		return nil, fmt.Errorf("GetProjectQuota error: %s", err)
	}

	if p.Code != 0 || p.Data == nil {
		return nil, fmt.Errorf("GetProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p, nil
}

// UpdateProjectQuota 更新项目资源额度
func UpdateProjectQuota(ctx context.Context, req *bcsproject.UpdateProjectQuotaRequest) (
	bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Quota.UpdateProjectQuota(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateProjectQuota error: %s", err)
	}

	if p.Code != 0 || p.Data == nil || p.Data.QuotaId == "" {
		return false, fmt.Errorf("UpdateProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}

// ScaleUpProjectQuota 扩展项目资源额度
func ScaleUpProjectQuota(ctx context.Context, req *bcsproject.ScaleUpProjectQuotaRequest) (bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Quota.ScaleUpProjectQuota(ctx, req)
	if err != nil {
		return false, fmt.Errorf("ScaleUpProjectQuota error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("ScaleUpProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}

// ScaleDownProjectQuota 缩减项目资源额度
func ScaleDownProjectQuota(ctx context.Context, req *bcsproject.ScaleDownProjectQuotaRequest) (bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Quota.ScaleDownProjectQuota(ctx, req)
	if err != nil {
		return false, fmt.Errorf("ScaleDownProjectQuota error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("ScaleDownProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}

// DeleteProjectQuota 删除项目资源额度
func DeleteProjectQuota(ctx context.Context, req *bcsproject.DeleteProjectQuotaRequest) (bool, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.Quota.DeleteProjectQuota(ctx, req)
	if err != nil {
		return false, fmt.Errorf("DeleteProjectQuota error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("DeleteProjectQuota error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return true, nil
}

// ListProjectQuotas 列出项目资源额度
func ListProjectQuotas(ctx context.Context, req *bcsproject.ListProjectQuotasRequest) (
	*bcsproject.ListProjectQuotasData, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Quota.ListProjectQuotas(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListProjectQuotas error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("ListProjectQuotas error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// ListProjectQuotasV2 列出项目资源额度
func ListProjectQuotasV2(ctx context.Context, req *bcsproject.ListProjectQuotasV2Request) (
	*bcsproject.ListProjectQuotasData, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Quota.ListProjectQuotasV2(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListProjectQuotasV2 error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("ListProjectQuotasV2 error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// GetProjectQuotasUsage 获取项目资源使用情况
func GetProjectQuotasUsage(ctx context.Context, quotaId string) (
	*bcsproject.GetProjectQuotasUsageData, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Quota.GetProjectQuotasUsage(ctx, &bcsproject.GetProjectQuotasUsageReq{
		QuotaId: quotaId,
	})
	if err != nil {
		return nil, fmt.Errorf("GetProjectQuotasUsage error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("GetProjectQuotasUsage error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}

// GetProjectQuotasStatistics 获取项目资源统计信息
func GetProjectQuotasStatistics(ctx context.Context, req *bcsproject.GetProjectQuotasStatisticsRequest) (
	*bcsproject.ProjectQuotasStatisticsData, error) {
	cli, close, err := bcsproject.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.Quota.GetProjectQuotasStatistics(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GetProjectQuotasStatistics error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("GetProjectQuotasStatistics error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}
