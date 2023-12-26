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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/template"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/templatespace"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/templateversion"
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

	// 模板文件文件夹
	GetTemplateSpace(ctx context.Context, id string) (*entity.TemplateSpace, error)
	ListTemplateSpace(ctx context.Context, cond *operator.Condition) ([]*entity.TemplateSpace, error)
	CreateTemplateSpace(ctx context.Context, templateSpace *entity.TemplateSpace) (string, error)
	UpdateTemplateSpace(ctx context.Context, id string, templateSpace entity.M) error
	DeleteTemplateSpace(ctx context.Context, id string) error

	// 模板文件元数据
	GetTemplate(ctx context.Context, id string) (*entity.Template, error)
	ListTemplate(ctx context.Context, cond *operator.Condition) ([]*entity.Template, error)
	CreateTemplate(ctx context.Context, template *entity.Template) (string, error)
	UpdateTemplate(ctx context.Context, id string, template entity.M) error
	UpdateTemplateBySpecial(
		ctx context.Context, projectCode, templateSpace string, template entity.M) error
	DeleteTemplate(ctx context.Context, id string) error
	DeleteTemplateBySpecial(ctx context.Context, projectCode, templateSpace string) error

	// 模板文件版本
	GetTemplateVersion(ctx context.Context, id string) (*entity.TemplateVersion, error)
	ListTemplateVersion(ctx context.Context, cond *operator.Condition) ([]*entity.TemplateVersion, error)
	CreateTemplateVersion(ctx context.Context, templateVersion *entity.TemplateVersion) (string, error)
	UpdateTemplateVersion(ctx context.Context, id string, templateVersion entity.M) error
	UpdateTemplateVersionBySpecial(
		ctx context.Context, projectCode, templateName, templateSpace string, templateVersion entity.M) error
	DeleteTemplateVersion(ctx context.Context, id string) error
	DeleteTemplateVersionBySpecial(ctx context.Context, projectCode, templateName, templateSpace string) error
}

type modelSet struct {
	*view.ModelView
	*templatespace.ModelTemplateSpace
	*template.ModelTemplate
	*templateversion.ModelTemplateVersion
}

// New return a new ClusterResourcesModel instance
func New(db drivers.DB) ClusterResourcesModel {
	return &modelSet{
		ModelView:            view.New(db),
		ModelTemplateSpace:   templatespace.New(db),
		ModelTemplate:        template.New(db),
		ModelTemplateVersion: templateversion.New(db),
	}
}
