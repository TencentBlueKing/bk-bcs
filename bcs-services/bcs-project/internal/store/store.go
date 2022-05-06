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

package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
)

type ProjectModel interface {
	CreateProject(ctx context.Context, project *project.Project) error
	GetProject(ctx context.Context, projectID string) (*project.Project, error)
	GetProjectByField(ctx context.Context, pf *project.ProjectField) (*project.Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	UpdateProject(ctx context.Context, project *project.Project) error
	ListProjects(ctx context.Context, cond *operator.Condition, opt *page.Pagination) ([]project.Project, int64, error)
	ListProjectByIDs(ctx context.Context, ids []string, opt *page.Pagination) ([]project.Project, int64, error)
}

type modelSet struct {
	*project.ModelProject
}

func New(db drivers.DB) ProjectModel {
	return &modelSet{
		ModelProject: project.New(db),
	}
}
