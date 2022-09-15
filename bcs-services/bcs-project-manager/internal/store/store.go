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

// Package store xxx
package store

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	vd "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vv "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
)

// ProjectModel project interface
type ProjectModel interface {
	CreateProject(ctx context.Context, project *project.Project) error
	GetProject(ctx context.Context, projectIDOrCode string) (*project.Project, error)
	GetProjectByField(ctx context.Context, pf *project.ProjectField) (*project.Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	UpdateProject(ctx context.Context, project *project.Project) error
	ListProjects(ctx context.Context, cond *operator.Condition, opt *page.Pagination) ([]project.Project, int64, error)
	ListProjectByIDs(ctx context.Context, ids []string, opt *page.Pagination) ([]project.Project, int64, error)

	CreateVariableDefinition(ctx context.Context, entity *vd.VariableDefinition) error
	UpdateVariableDefinition(ctx context.Context, entity entity.M) (*vd.VariableDefinition, error)
	UpsertVariableDefinition(ctx context.Context, entity *vd.VariableDefinition) error
	GetVariableDefinition(ctx context.Context, variableID string) (*vd.VariableDefinition, error)
	GetVariableDefinitionByKey(ctx context.Context, projectCode, key string) (*vd.VariableDefinition, error)
	ListVariableDefinitions(ctx context.Context,
		cond *operator.Condition, opt *page.Pagination) ([]vd.VariableDefinition, int64, error)
	DeleteVariableDefinitions(ctx context.Context, ids []string) (int64, error)

	CreateVariableValue(ctx context.Context, vv *vv.VariableValue) error
	GetVariableValue(ctx context.Context,
		variableID, clusterID, namespace, scope string) (*vv.VariableValue, error)
	UpsertVariableValue(ctx context.Context,
		value *vv.VariableValue) error
}

type modelSet struct {
	*project.ModelProject
	*vd.ModelVariableDefinition
	*vv.ModelVariableValue
}

var model *modelSet

// New new project model
func New(db drivers.DB) ProjectModel {
	return &modelSet{
		ModelProject:            project.New(db),
		ModelVariableDefinition: vd.New(db),
		ModelVariableValue:      vv.New(db),
	}
}

// InitModel init model
func InitModel(db drivers.DB) {
	model = &modelSet{
		ModelProject:            project.New(db),
		ModelVariableDefinition: vd.New(db),
		ModelVariableValue:      vv.New(db),
	}
}

// GetModel get model
func GetModel() ProjectModel {
	return model
}
