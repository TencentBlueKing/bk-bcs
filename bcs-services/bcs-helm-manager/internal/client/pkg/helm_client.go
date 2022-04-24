/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package pkg

import (
	"context"

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
	List(ctx context.Context, req *helmmanager.ListRepositoryReq) (*helmmanager.RepositoryListData, error)
}

// ChartClient define the chart operation handler
type ChartClient interface {
	List(ctx context.Context, req *helmmanager.ListChartReq) (*helmmanager.ChartListData, error)
	Versions(ctx context.Context, req *helmmanager.ListChartVersionReq) (*helmmanager.ChartVersionListData, error)
	Detail(ctx context.Context, req *helmmanager.GetChartDetailReq) (*helmmanager.ChartDetail, error)
}

// ReleaseClient define the release operation handler
type ReleaseClient interface {
	List(ctx context.Context, req *helmmanager.ListReleaseReq) (*helmmanager.ReleaseListData, error)
	Install(ctx context.Context, req *helmmanager.InstallReleaseReq) (*helmmanager.ReleaseDetail, error)
	Uninstall(ctx context.Context, req *helmmanager.UninstallReleaseReq) error
	Upgrade(ctx context.Context, req *helmmanager.UpgradeReleaseReq) (*helmmanager.ReleaseDetail, error)
	Rollback(ctx context.Context, req *helmmanager.RollbackReleaseReq) error
}
