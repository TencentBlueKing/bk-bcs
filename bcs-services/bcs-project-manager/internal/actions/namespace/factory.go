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

// Package namespace xxx
package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/shared"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

// NamespaceFactory namespace faction factory
// nolint
type NamespaceFactory struct {
	model store.ProjectModel
}

// NewNamespaceFactory new namespace faction factory
func NewNamespaceFactory(model store.ProjectModel) *NamespaceFactory {
	return &NamespaceFactory{
		model: model,
	}
}

// Action get action by clusterID
func (f *NamespaceFactory) Action(ctx context.Context, clusterID, projectIDOrCode string) (
	action.NamespaceAction, error) {
	cluster, err := clustermanager.GetCluster(ctx, clusterID, true)
	if err != nil {
		logging.Error("get cluster %s from cluster-manager failed, err: %s", cluster, err.Error())
		return nil, err
	}
	// projectIDOrCode 为 '-' 则不校验项目信息
	if projectIDOrCode == "-" {
		if cluster.GetIsShared() {
			return shared.NewSharedNamespaceAction(f.model), nil
		}
		return independent.NewIndependentNamespaceAction(f.model), nil
	}
	project, err := f.model.GetProject(context.TODO(), projectIDOrCode)
	if err != nil {
		logging.Error("get project from db failed, err: %s", err.Error())
		return nil, err
	}
	if cluster.GetProjectID() != project.ProjectID {
		if cluster.GetIsShared() {
			return shared.NewSharedNamespaceAction(f.model), nil
		}
		return nil, errorx.NewReadableErr(errorx.ParamErr, "project or cluster not valid")
	}
	return independent.NewIndependentNamespaceAction(f.model), nil
}
