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

// Package store xxx
package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/envmanage"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/utils"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/view"
)

// ClusterResourcesModel 提供了一套完整的 cluster-resources 所需的数据库操作接口
type ClusterResourcesModel interface {
	// 视图配置管理
	CreateView(ctx context.Context, view *entity.View) (string, error)
	UpdateView(ctx context.Context, id string, view entity.M) error
	GetView(ctx context.Context, id string) (*entity.View, error)
	GetViewByName(ctx context.Context, projectCode, name string) (*entity.View, error)
	ListViews(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (int64, []*entity.View, error)
	DeleteView(ctx context.Context, id string) error

	// 环境管理
	GetEnvManage(ctx context.Context, id string) (*entity.EnvManage, error)
	ListEnvManages(ctx context.Context, cond *operator.Condition) ([]*entity.EnvManage, error)
	CreateEnvManage(ctx context.Context, envManage *entity.EnvManage) (string, error)
	UpdateEnvManage(ctx context.Context, id string, envManage entity.M) error
	DeleteEnvManage(ctx context.Context, id string) error
}

type modelSet struct {
	*view.ModelView
	*envmanage.ModelEnvManage
}

// New return a new ClusterResourcesModel instance
func New(db drivers.DB) ClusterResourcesModel {
	return &modelSet{
		ModelView:      view.New(db),
		ModelEnvManage: envmanage.New(db),
	}
}
