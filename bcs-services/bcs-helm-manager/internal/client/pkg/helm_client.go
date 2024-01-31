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

// Package pkg xxx
package pkg

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/utils/chart"
	helmmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/proto/bcs-helm-manager"
)

// HelmClient define the whole operation for helm manager
type HelmClient interface {
	Available(ctx context.Context) error
	Repository() RepositoryClient
	Chart() ChartClient
	Release() ReleaseClient
}

// RepositoryClient define the repository operation handler
type RepositoryClient interface {
	Create(ctx context.Context, req *helmmanager.CreateRepositoryReq) error
	Update(ctx context.Context, req *helmmanager.UpdateRepositoryReq) error
	Delete(ctx context.Context, req *helmmanager.DeleteRepositoryReq) error
	Get(ctx context.Context, req *helmmanager.GetRepositoryReq) (*helmmanager.Repository, error)
	List(ctx context.Context, req *helmmanager.ListRepositoryReq) ([]*helmmanager.Repository, error)
}

// ChartClient define the chart operation handler
type ChartClient interface {
	List(ctx context.Context, req *helmmanager.ListChartV1Req) (*helmmanager.ChartListData, error)
	GetChartDetail(ctx context.Context, req *helmmanager.GetChartDetailV1Req) (*helmmanager.Chart, error)
	Versions(ctx context.Context, req *helmmanager.ListChartVersionV1Req) (*helmmanager.ChartVersionListData, error)
	GetVersionDetail(ctx context.Context, req *helmmanager.GetVersionDetailV1Req) (*helmmanager.ChartDetail, error)
	DeleteChart(ctx context.Context, req *helmmanager.DeleteChartReq) error
	DeleteChartVersion(ctx context.Context, req *helmmanager.DeleteChartVersionReq) error
	PushChart(ctx context.Context, req *chart.PushChart) error
}

// ReleaseClient define the release operation handler
type ReleaseClient interface {
	GetReleaseDetail(ctx context.Context, req *helmmanager.GetReleaseDetailV1Req) (*helmmanager.ReleaseDetail, error)
	List(ctx context.Context, req *helmmanager.ListReleaseV1Req) (*helmmanager.ReleaseListData, error)
	Install(ctx context.Context, req *helmmanager.InstallReleaseV1Req) error
	Uninstall(ctx context.Context, req *helmmanager.UninstallReleaseV1Req) error
	Upgrade(ctx context.Context, req *helmmanager.UpgradeReleaseV1Req) error
	Rollback(ctx context.Context, req *helmmanager.RollbackReleaseV1Req) error
	GetReleaseHistory(ctx context.Context, req *helmmanager.GetReleaseHistoryReq) ([]*helmmanager.ReleaseHistory, error)
}
