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

package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/release"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/repository"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
)

// HelmManagerModel 提供了一套完整的helm manager所需的数据库操作接口
type HelmManagerModel interface {
	// CreateRepository 创建仓库
	CreateRepository(ctx context.Context, repository *entity.Repository) error

	// UpdateRepository 更新仓库, 主键为projectID + name
	UpdateRepository(ctx context.Context, projectID, name string, repository entity.M) error

	// GetRepository 根据主键查询仓库信息
	GetRepository(ctx context.Context, projectID, name string) (*entity.Repository, error)

	// ListRepository 根据条件查询仓库列表
	// 其中分页配置详见 utils.ListOption, 采用 page + size 的模式
	ListRepository(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
		int64, []*entity.Repository, error)

	// DeleteRepository 根据主键删除仓库
	DeleteRepository(ctx context.Context, projectID, name string) error

	// DeleteRepositories 根据给定的projectID和name列表, 批量删除一批仓库
	DeleteRepositories(ctx context.Context, projectID string, names []string) error

	// CreateRelease 创建一个release
	CreateRelease(ctx context.Context, release *entity.Release) error

	// GetRelease 精确到revision, 获取一个release
	GetRelease(ctx context.Context, clusterID, namespace, name string, revision int) (*entity.Release, error)

	// ListRelease 根据条件查询仓库列表
	// 其中分页配置详见 utils.ListOption, 采用 page + size 的模式
	ListRelease(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (int64, []*entity.Release, error)

	// DeleteRelease 删除对应revision的release
	DeleteRelease(ctx context.Context, clusterID, namespace, name string, revision int) error

	// DeleteReleases 删除指定clusterID-namespace-name下的所有revision
	DeleteReleases(ctx context.Context, clusterID, namespace, name string) error
}

type modelSet struct {
	*repository.ModelRepository
	*release.ModelRelease
}

// New return a new ResourceManagerModel instance
func New(db drivers.DB) HelmManagerModel {
	return &modelSet{
		ModelRepository: repository.New(db),
		ModelRelease:    release.New(db),
	}
}
