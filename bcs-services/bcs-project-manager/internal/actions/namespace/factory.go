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

package namespace

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/independent"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/namespace/shared"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
)

//NamespaceFactory namespace faction factory
type NamespaceFactory struct {
	model store.ProjectModel
}

//NewNamespaceFactory new namespace faction factory
func NewNamespaceFactory(model store.ProjectModel) *NamespaceFactory {
	return &NamespaceFactory{
		model: model,
	}
}

// Action get action by clusterID
func (f *NamespaceFactory) Action(clusterID string) (action.NamespaceAction, error) {
	cli, closeCon, err := clustermanager.GetClusterManagerClient()
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &clustermanager.GetClusterReq{
		ClusterID: clusterID,
	}
	resp, err := cli.GetCluster(context.Background(), req)
	if err != nil {
		logging.Error("list cluster from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetData().GetIsShared() {
		return shared.NewSharedNamespaceAction(f.model), nil
	}
	return independent.NewIndependentNamespaceAction(f.model), nil
}
