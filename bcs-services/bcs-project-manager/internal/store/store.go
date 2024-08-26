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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/config"
	nsm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	vdm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variabledefinition"
	vvm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/variablevalue"
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
	SearchProjects(ctx context.Context, ids, limitIDs []string, searchKey, kind string, opt *page.Pagination) (
		[]project.Project, int64, error)

	GetNamespace(ctx context.Context, projectCode, clusterID, name string) (*nsm.Namespace, error)
	GetNamespaceByItsmTicketType(ctx context.Context,
		projectCode, clusterID, namespace, stagingType string) (*nsm.Namespace, error)
	CreateNamespace(ctx context.Context, ns *nsm.Namespace) error
	ListNamespaces(ctx context.Context) ([]nsm.Namespace, error)
	ListNamespacesByItsmTicketType(ctx context.Context,
		projectCode, clusterID string, types []string) ([]nsm.Namespace, error)
	UpdateNamespace(ctx context.Context, ns entity.M) (*nsm.Namespace, error)
	DeleteNamespace(ctx context.Context, projectCode, clusterID, namespace string) error

	CreateVariableDefinition(ctx context.Context, entity *vdm.VariableDefinition) error
	UpdateVariableDefinition(ctx context.Context, entity entity.M) (*vdm.VariableDefinition, error)
	UpsertVariableDefinition(ctx context.Context, entity *vdm.VariableDefinition) error
	GetVariableDefinition(ctx context.Context, variableID string) (*vdm.VariableDefinition, error)
	GetVariableDefinitionByKey(ctx context.Context, projectCode, key string) (*vdm.VariableDefinition, error)
	ListVariableDefinitions(ctx context.Context,
		cond *operator.Condition, opt *page.Pagination) ([]vdm.VariableDefinition, int64, error)
	DeleteVariableDefinitions(ctx context.Context, ids []string) (int64, error)

	CreateVariableValue(ctx context.Context, vv *vvm.VariableValue) error
	GetVariableValue(ctx context.Context, variableID, clusterID, namespace, scope string) (*vvm.VariableValue, error)
	UpsertVariableValue(ctx context.Context, value *vvm.VariableValue) error
	ListClusterVariableValues(ctx context.Context, variableID string) ([]vvm.VariableValue, error)
	ListNamespaceVariableValues(ctx context.Context, variableID, clusterID string) ([]vvm.VariableValue, error)
	ListVariableValuesInCluster(ctx context.Context, clusterID string) ([]vvm.VariableValue, error)
	ListVariableValuesInNamespace(ctx context.Context, clusterID, namespace string) ([]vvm.VariableValue, error)
	ListVariableValuesInAllNamespace(ctx context.Context, clusterID string) ([]vvm.VariableValue, error)
	DeleteVariableValuesByNamespace(ctx context.Context, clusterID, namespace string) (int64, error)

	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key, value string) error
}

type modelSet struct {
	*project.ModelProject
	*nsm.ModelNamespace
	*vdm.ModelVariableDefinition
	*vvm.ModelVariableValue
	*config.ModelConfig
}

var model *modelSet

// New new project model
func New(db drivers.DB) ProjectModel {
	return &modelSet{
		ModelProject:            project.New(db),
		ModelNamespace:          nsm.New(db),
		ModelVariableDefinition: vdm.New(db),
		ModelVariableValue:      vvm.New(db),
		ModelConfig:             config.New(db),
	}
}

// InitModel init model
func InitModel(db drivers.DB) {
	model = &modelSet{
		ModelProject:            project.New(db),
		ModelNamespace:          nsm.New(db),
		ModelVariableDefinition: vdm.New(db),
		ModelVariableValue:      vvm.New(db),
		ModelConfig:             config.New(db),
	}
}

// GetModel get model
func GetModel() ProjectModel {
	return model
}
